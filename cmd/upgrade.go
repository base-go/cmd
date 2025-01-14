package cmd

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

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

func getLatestRelease() (*Release, error) {
	url := "https://api.github.com/repos/base-go/cmd/releases/latest"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
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

		if header.Typeflag == tar.TypeReg {
			outFile, err := os.Create(targetPath)
			if err != nil {
				return err
			}
			defer outFile.Close()

			if _, err := io.Copy(outFile, tarReader); err != nil {
				return err
			}
			return os.Chmod(targetPath, 0755)
		}
	}
	return nil
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

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade Base CLI to the latest version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Checking for updates...")

		release, err := getLatestRelease()
		if err != nil {
			fmt.Printf("Error checking for updates: %v\n", err)
			return
		}

		currentVersion := "1.0.0" // Replace with actual current version
		latestVersion := strings.TrimPrefix(release.TagName, "v")

		if currentVersion == latestVersion {
			fmt.Printf("You're already using the latest version (%s)\n", currentVersion)
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

		// Get the current executable path
		execPath, err := os.Executable()
		if err != nil {
			fmt.Printf("Error getting executable path: %v\n", err)
			return
		}

		// Create a temporary file for the binary
		tmpDir := os.TempDir()
		tmpFile := filepath.Join(tmpDir, "base-new")

		// Download and extract the new version
		if err := downloadAndInstall(downloadURL, tmpFile); err != nil {
			fmt.Printf("Error downloading update: %v\n", err)
			return
		}

		// Replace the old binary
		if err := os.Rename(tmpFile, execPath); err != nil {
			// If direct rename fails (e.g., on Windows), try copy and remove
			if err := copyFile(tmpFile, execPath); err != nil {
				fmt.Printf("Error installing update: %v\n", err)
				os.Remove(tmpFile)
				return
			}
			os.Remove(tmpFile)
		}

		fmt.Printf("Successfully upgraded to version %s!\n", latestVersion)
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
