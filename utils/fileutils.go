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

// UpdateInitGo adds the module initialization to init.go
func UpdateInitGo(packageName, structName string) error {
	initFile := filepath.Join("app", "init.go")
	if _, err := os.Stat(initFile); err != nil {
		return fmt.Errorf("init.go not found: %v", err)
	}

	content, err := os.ReadFile(initFile)
	if err != nil {
		return fmt.Errorf("error reading init.go: %v", err)
	}

	// Add import if not exists
	importPath := fmt.Sprintf(`"base/app/%s"`, packageName)
	contentStr := string(content)
	if !strings.Contains(contentStr, importPath) {
		importMarker := "// MODULE_IMPORT_MARKER"
		markerIndex := strings.Index(contentStr, importMarker)
		if markerIndex == -1 {
			return fmt.Errorf("could not find import marker comment in init.go")
		}
		// Find the end of the line containing the marker
		lineEnd := strings.Index(contentStr[markerIndex:], "\n") + markerIndex
		if lineEnd == -1 {
			lineEnd = len(contentStr)
		}
		newContent := contentStr[:lineEnd+1] + "\t" + importPath + "\n" + contentStr[lineEnd+1:]
		contentStr = newContent
	}

	// Create the module initializer line
	moduleInitializer := fmt.Sprintf(`	modules["%s"] = %s.New%sModule(deps.DB)
`, packageName, packageName, structName)

	// Check if the module initializer already exists
	if strings.Contains(contentStr, fmt.Sprintf(`modules["%s"]`, packageName)) {
		fmt.Printf("Module initializer for %s already exists in init.go\n", packageName)
		return nil
	}

	// Insert the initializer before the marker comment
	markerComment := "// MODULE_INITIALIZER_MARKER"
	markerIndex := strings.Index(contentStr, markerComment)
	if markerIndex == -1 {
		return fmt.Errorf("could not find marker comment in init.go")
	}

	// Find the start of the line containing the marker
	lineStart := strings.LastIndex(contentStr[:markerIndex], "\n") + 1

	// Split the content at the marker line
	beforeMarker := contentStr[:lineStart]
	afterMarker := contentStr[lineStart:]

	// Combine all parts
	newContent := beforeMarker + moduleInitializer + afterMarker

	if err := os.WriteFile(initFile, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("error writing to init.go: %v", err)
	}

	return nil
}
