package version

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Version information
var (
	// Version is the current version of Base CLI
	Version = "2.0.0"

	// CommitHash is the git commit hash at build time
	CommitHash = "unknown"

	// BuildDate is the date when the binary was built
	BuildDate = "unknown"

	// GoVersion is the version of Go used to build the binary
	GoVersion = "unknown"
)

// BuildInfo contains all version information
type BuildInfo struct {
	Version    string `json:"version"`
	CommitHash string `json:"commit_hash"`
	BuildDate  string `json:"build_date"`
	GoVersion  string `json:"go_version"`
}

// Release represents a GitHub release
type Release struct {
	TagName     string    `json:"tag_name"`
	PublishedAt time.Time `json:"published_at"`
	Body        string    `json:"body"`
	HTMLURL     string    `json:"html_url"`
	Assets      []Asset   `json:"assets"`
}

// Asset represents a release asset
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// GetBuildInfo returns all version information
func GetBuildInfo() BuildInfo {
	return BuildInfo{
		Version:    Version,
		CommitHash: CommitHash,
		BuildDate:  BuildDate,
		GoVersion:  GoVersion,
	}
}

// String returns a string representation of version information
func (bi BuildInfo) String() string {
	return fmt.Sprintf("Base CLI %s\nCommit: %s\nBuilt: %s\nGo version: %s",
		bi.Version, bi.CommitHash, bi.BuildDate, bi.GoVersion)
}

// CheckLatestVersion checks GitHub for newer releases
func CheckLatestVersion() (*Release, error) {
	url := "https://api.github.com/repos/base-go/cmd/releases/latest"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var release Release
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, err
	}

	return &release, nil
}

// HasUpdate checks if the current version is behind the latest release
func HasUpdate(current, latest string) bool {
	if current == "dev" {
		return true
	}
	// Normalize versions by removing 'v' prefix if present
	current = strings.TrimPrefix(current, "v")
	latest = strings.TrimPrefix(latest, "v")
	// If versions are equal, there's no update
	if current == latest {
		return false
	}
	// Compare versions (you might want to add semantic version comparison here)
	return current != latest
}

// GetAllReleases gets all releases from GitHub
func GetAllReleases() ([]Release, error) {
	url := "https://api.github.com/repos/base-go/cmd/releases"
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var releases []Release
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, err
	}

	return releases, nil
}

// CompareVersions compares two semantic versions
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func CompareVersions(v1, v2 string) int {
	// Remove 'v' prefix if present
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")
	
	if v1 == v2 {
		return 0
	}
	
	// Split versions into parts
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")
	
	// Pad shorter version with zeros
	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}
	
	for len(parts1) < maxLen {
		parts1 = append(parts1, "0")
	}
	for len(parts2) < maxLen {
		parts2 = append(parts2, "0")
	}
	
	// Compare each part numerically
	for i := 0; i < maxLen; i++ {
		// Convert to integers for proper numeric comparison
		var num1, num2 int
		fmt.Sscanf(parts1[i], "%d", &num1)
		fmt.Sscanf(parts2[i], "%d", &num2)
		
		if num1 < num2 {
			return -1
		}
		if num1 > num2 {
			return 1
		}
	}
	
	return 0
}

// FormatUpdateMessage returns a formatted update message
func FormatUpdateMessage(current, latest, releaseURL, releaseNotes string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n📦 Update available! %s → %s\n", current, latest))
	sb.WriteString("Run: base upgrade\n")
	sb.WriteString(fmt.Sprintf("Release notes: %s\n", releaseURL))
	if releaseNotes != "" {
		sb.WriteString(fmt.Sprintf("\nChangelog:\n%s\n", releaseNotes))
	}
	return sb.String()
}
