package version

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// VersionCheckURL is the URL to check for latest commit
	VersionCheckURL = "https://api.github.com/repos/base-go/cmd/commits/main"
)

// CommitInfo represents the GitHub commit response
type CommitInfo struct {
	SHA    string `json:"sha"`
	Commit struct {
		Message string `json:"message"`
		Author  struct {
			Date string `json:"date"`
		} `json:"author"`
	} `json:"commit"`
	HTMLURL string `json:"html_url"`
}

// CheckForUpdates checks if there's a newer version available
func CheckForUpdates() (string, string, bool, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(VersionCheckURL)
	if err != nil {
		return "", "", false, nil // Silently fail on network errors
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", false, nil
	}

	var commit CommitInfo
	if err := json.Unmarshal(body, &commit); err != nil {
		return "", "", false, nil
	}

	// Compare commit hashes
	if commit.SHA != CommitHash {
		return commit.SHA[:8], commit.HTMLURL, true, nil
	}

	return commit.SHA[:8], commit.HTMLURL, false, nil
}

// PrintUpdateMessage prints update message if a new version is available
func PrintUpdateMessage() {
	latestCommit, url, hasUpdate, err := CheckForUpdates()
	if err != nil || !hasUpdate {
		return
	}

	fmt.Printf("\n📦 Update available! %s → %s\n", CommitHash[:8], latestCommit)
	fmt.Printf("Run: base upgrade\n")
	fmt.Printf("Latest changes: %s\n\n", url)
}
