package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/mutablelogic/go-client/pkg/version"
)

type VersionCmd struct{}

func (v *VersionCmd) Run(globals *Globals) error {
	w := os.Stdout
	if version.GitSource != "" {
		if version.GitTag != "" {
			fmt.Fprintf(w, " %v", version.GitTag)
		}
		if version.GitSource != "" {
			fmt.Fprintf(w, " (%v)", version.GitSource)
		}
		fmt.Fprintln(w, "")
	}
	if runtime.Version() != "" {
		fmt.Fprintf(w, "%v %v/%v\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	}
	if version.GitBranch != "" {
		fmt.Fprintf(w, "Branch: %v\n", version.GitBranch)
	}
	if version.GitHash != "" {
		fmt.Fprintf(w, "Hash: %v\n", version.GitHash)
	}
	if version.GoBuildTime != "" {
		fmt.Fprintf(w, "BuildTime: %v\n", version.GoBuildTime)
	}
	return nil
}
