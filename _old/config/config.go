package config

import (
	"fmt"
	"io"
	"runtime"
)

///////////////////////////////////////////////////////////////////////////////
// GLOBALS

var (
	GitSource   string
	GitTag      string
	GitBranch   string
	GitHash     string
	GoBuildTime string
)

func PrintVersion(w io.Writer) {
	if GitSource != "" {
		fmt.Fprintf(w, "  URL: https://%v\n", GitSource)
	}
	if GitTag != "" || GitBranch != "" {
		fmt.Fprintf(w, "  Version: %v (branch: %q hash:%q)\n", GitTag, GitBranch, GitHash)
	}
	if GoBuildTime != "" {
		fmt.Fprintf(w, "  Build Time: %v\n", GoBuildTime)
	}
	fmt.Fprintf(w, "  Go: %v (%v/%v)\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
}
