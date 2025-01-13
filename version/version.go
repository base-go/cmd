package version

// Version information
var (
	// Version is the current version of Base CLI
	Version string

	// CommitHash is the git commit hash at build time
	CommitHash string

	// BuildDate is the date when the binary was built
	BuildDate string

	// GoVersion is the version of Go used to build the binary
	GoVersion string
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
	if Version == "" {
		Version = "1.0.0"
	}
	if CommitHash == "" {
		CommitHash = "unknown"
	}
	if BuildDate == "" {
		BuildDate = "unknown"
	}
	if GoVersion == "" {
		GoVersion = "unknown"
	}
	
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
