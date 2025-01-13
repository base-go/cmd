package version

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

// GithubRelease represents a GitHub release
type GithubRelease struct {
	TagName     string    `json:"tag_name"`
	PublishedAt time.Time `json:"published_at"`
	Body        string    `json:"body"`
	HTMLURL     string    `json:"html_url"`
}

// CheckForUpdates checks GitHub for newer releases
func CheckForUpdates() (latestVersion, releaseURL, releaseNotes string, hasUpdate bool, err error) {
	url := "https://api.github.com/repos/base-go/cmd/releases/latest"
	resp, err := http.Get(url)
	if err != nil {
		return "", "", "", false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", "", false, err
	}

	var release GithubRelease
	if err := json.Unmarshal(body, &release); err != nil {
		return "", "", "", false, err
	}

	latestVersion = strings.TrimPrefix(release.TagName, "v")
	hasUpdate = Version == "dev" || Version != latestVersion

	return latestVersion, release.HTMLURL, release.Body, hasUpdate, nil
}

// CheckForUpdates checks if there's a newer version available
func CheckForUpdatesCommit() (string, string, bool, error) {
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
	latestVersion, releaseURL, releaseNotes, hasUpdate, err := CheckForUpdates()
	if err != nil || !hasUpdate {
		return
	}

	currentVersion := Version
	if currentVersion == "unknown" || currentVersion == "" {
		currentVersion = "dev"
	}

	fmt.Printf("\n📦 Update available! %s → %s\n", currentVersion, latestVersion)
	fmt.Printf("Run: base upgrade\n")
	fmt.Printf("Release notes: %s\n", releaseURL)
	if releaseNotes != "" {
		fmt.Printf("\nChangelog:\n%s\n", releaseNotes)
	}
}
