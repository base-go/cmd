package version

// Version information
var (
	// Version is the current version of Base CLI
	Version = "1.0.0"

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
	return "Base CLI " + bi.Version + "\n" +
		"Commit: " + bi.CommitHash + "\n" +
		"Built: " + bi.BuildDate + "\n" +
		"Go version: " + bi.GoVersion
}
