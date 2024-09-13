package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Unzip extracts a zip archive to a destination directory.
func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		// Ensure that the file path is within the destination directory.
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("invalid file path: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			// Create directory if it doesn't exist.
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Create parent directories if necessary.
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		// Create the file.
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		// Open the file inside the zip archive.
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		// Copy the file content to the destination file.
		_, err = io.Copy(outFile, rc)

		// Close files.
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}
