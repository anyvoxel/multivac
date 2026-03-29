// Package version provides build-time version information.
package version

import "runtime"

// These variables are intended to be set via -ldflags "-X ..." at build time.
var (
	module       = "multivac"
	version      = "dev"
	branch       = ""
	gitCommit    = ""
	gitTreeState = ""
	buildDate    = ""
)

// BuildInfo is a stable shape for /version.
type BuildInfo struct {
	Module       string `json:"module"`
	Version      string `json:"version"`
	Branch       string `json:"branch,omitempty"`
	GitCommit    string `json:"gitCommit,omitempty"`
	GitTreeState string `json:"gitTreeState,omitempty"`
	BuildDate    string `json:"buildDate,omitempty"`
	GoVersion    string `json:"goVersion"`
}

// Info returns the current build information.
func Info() BuildInfo {
	return BuildInfo{
		Module:       module,
		Version:      version,
		Branch:       branch,
		GitCommit:    gitCommit,
		GitTreeState: gitTreeState,
		BuildDate:    buildDate,
		GoVersion:    runtime.Version(),
	}
}
