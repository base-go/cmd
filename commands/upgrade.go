package cmd

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/BaseTechStack/basecmd/version"
	"github.com/spf13/cobra"
)

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

func extractTarGz(gzipStream io.Reader, targetPath string) error {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return err
	}
	defer uncompressedStream.Close()

	tarReader := tar.NewReader(uncompressedStream)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// We're looking for the binary file which should be named "base" or "base.exe"
		if header.Typeflag == tar.TypeReg {
			baseName := filepath.Base(header.Name)
			expectedName := "base"
			if runtime.GOOS == "windows" {
				expectedName = "base.exe"
			}

			if baseName == expectedName {
				outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
				if err != nil {
					return err
				}
				defer outFile.Close()

				if _, err := io.Copy(outFile, tarReader); err != nil {
					return err
				}
				return nil
			}
		}
	}
	return fmt.Errorf("binary not found in archive")
}

func extractZip(zipFile *os.File, targetPath string) error {
	stat, err := zipFile.Stat()
	if err != nil {
		return err
	}

	reader, err := zip.NewReader(zipFile, stat.Size())
	if err != nil {
		return err
	}

	for _, file := range reader.File {
		if !file.FileInfo().IsDir() {
			rc, err := file.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			outFile, err := os.Create(targetPath)
			if err != nil {
				return err
			}
			defer outFile.Close()

			if _, err := io.Copy(outFile, rc); err != nil {
				return err
			}
			return os.Chmod(targetPath, 0755)
		}
	}
	return nil
}

func downloadAndInstall(url, targetPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Create a temporary file for the archive
	tmpDir := os.TempDir()
	tmpArchive := filepath.Join(tmpDir, "base-archive")
	out, err := os.Create(tmpArchive)
	if err != nil {
		return err
	}
	defer func() {
		out.Close()
		os.Remove(tmpArchive)
	}()

	if _, err = io.Copy(out, resp.Body); err != nil {
		return err
	}
	out.Close()

	// Reopen the file for reading
	archiveFile, err := os.Open(tmpArchive)
	if err != nil {
		return err
	}
	defer archiveFile.Close()

	// Extract based on file type
	if strings.HasSuffix(url, ".zip") {
		return extractZip(archiveFile, targetPath)
	} else if strings.HasSuffix(url, ".tar.gz") {
		return extractTarGz(archiveFile, targetPath)
	}

	return fmt.Errorf("unsupported archive format")
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return os.Chmod(dst, 0755)
}

func runWithSudo(command string, args ...string) error {
	cmd := exec.Command("sudo", append([]string{command}, args...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade Base CLI to the latest version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Checking for updates...")

		release, err := version.CheckLatestVersion()
		if err != nil {
			fmt.Printf("Error checking for updates: %v\n", err)
			return
		}

		info := version.GetBuildInfo()
		latestVersion := strings.TrimPrefix(release.TagName, "v")

		if info.Version == latestVersion {
			fmt.Printf("You're already using the latest version (%s)\n", info.Version)
			return
		}

		// Determine the correct asset name based on OS and architecture
		osName := runtime.GOOS
		archName := runtime.GOARCH
		assetPrefix := fmt.Sprintf("base_%s_%s", osName, archName)
		var assetSuffix string
		if osName == "windows" {
			assetSuffix = ".zip"
		} else {
			assetSuffix = ".tar.gz"
		}

		var downloadURL string
		for _, asset := range release.Assets {
			if strings.HasPrefix(asset.Name, assetPrefix) && strings.HasSuffix(asset.Name, assetSuffix) {
				downloadURL = asset.BrowserDownloadURL
				break
			}
		}

		if downloadURL == "" {
			fmt.Printf("No compatible binary found for your system (%s_%s)\n", osName, archName)
			return
		}

		fmt.Printf("Downloading version %s...\n", latestVersion)

		// Create a temporary file for the binary
		tmpDir := os.TempDir()
		tmpFile := filepath.Join(tmpDir, "base-new")

		// Download and extract the new version
		if err := downloadAndInstall(downloadURL, tmpFile); err != nil {
			fmt.Printf("Error downloading update: %v\n", err)
			return
		}

		// Verify the binary is executable and matches our architecture
		execCmd := exec.Command(tmpFile, "version")
		if err := execCmd.Run(); err != nil {
			fmt.Printf("Error verifying binary: %v\n", err)
			os.Remove(tmpFile)
			return
		}

		// Make the temporary file executable
		if err := os.Chmod(tmpFile, 0755); err != nil {
			fmt.Printf("Error making binary executable: %v\n", err)
			os.Remove(tmpFile)
			return
		}

		// Get the installation directory and binary name based on OS
		var installDir, binDir, binaryName string
		if runtime.GOOS == "windows" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				fmt.Printf("Error getting home directory: %v\n", err)
				os.Remove(tmpFile)
				return
			}
			installDir = filepath.Join(homeDir, ".base")
			binDir = filepath.Join(homeDir, "bin")
			binaryName = "base.exe"
		} else {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				fmt.Printf("Error getting home directory: %v\n", err)
				os.Remove(tmpFile)
				return
			}
			installDir = filepath.Join(homeDir, ".base")
			binDir = "/usr/local/bin"
			binaryName = "base"
		}

		// Create installation directories
		if err := os.MkdirAll(installDir, 0755); err != nil {
			fmt.Printf("Error creating installation directory: %v\n", err)
			os.Remove(tmpFile)
			return
		}

		if err := os.MkdirAll(binDir, 0755); err != nil && !os.IsExist(err) {
			fmt.Printf("Error: Unable to create %s directory. Please run with sudo\n", binDir)
			os.Remove(tmpFile)
			return
		}

		destPath := filepath.Join(installDir, binaryName)

		// Move the new binary to installation directory
		if err := os.Rename(tmpFile, destPath); err != nil {
			fmt.Printf("Error moving binary: %v\n", err)
			os.Remove(tmpFile)
			return
		}

		// Create symlink or copy based on OS
		if runtime.GOOS == "windows" {
			// On Windows, copy to bin directory
			if err := copyFile(destPath, filepath.Join(binDir, binaryName)); err != nil {
				fmt.Printf("Error copying binary to bin directory: %v\n", err)
				return
			}
		} else {
			// On Unix systems, create symlink with sudo
			fmt.Println("Creating symlink in /usr/local/bin (requires sudo)...")
			if err := runWithSudo("ln", "-sf", destPath, filepath.Join(binDir, binaryName)); err != nil {
				fmt.Printf("Error updating symlink. Please run manually:\nsudo ln -sf %s %s\n",
					destPath, filepath.Join(binDir, binaryName))
				return
			}
		}

		fmt.Printf("Successfully upgraded to version %s!\n", latestVersion)
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
