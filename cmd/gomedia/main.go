package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	// Packages
	cmd "github.com/mutablelogic/go-server/pkg/cmd"
	httpresponse "github.com/mutablelogic/go-server/pkg/httpresponse"
	version "github.com/mutablelogic/go-server/pkg/version"
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

func main() {
	if err := cmd.Main(CLI{}, "gomedia is a client server application for media management", version.Version()); err != nil {
		str, code := formatError(err)
		fmt.Fprintln(os.Stderr, "Error:", str)
		os.Exit(code)
	}
}

func formatError(err error) (string, int) {
	var errResponse httpresponse.ErrResponse
	if errors.As(err, &errResponse) {
		if reason := strings.TrimSpace(errResponse.Reason); reason != "" {
			return reason, errResponse.Code
		}
	}
	return err.Error(), -1
}
