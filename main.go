package main

import (
	"fmt"
	"os"

	"github.com/base-go/cmd/cmd"
	"github.com/base-go/cmd/version"
)

// Version information set by build flags
var (
	Version    = "1.0.11"
	CommitHash = "unknown"
	BuildDate  = "unknown"
	GoVersion  = "unknown"
)

func init() {
	version.Version = Version
	version.CommitHash = CommitHash
	version.BuildDate = BuildDate
	version.GoVersion = GoVersion
}

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
