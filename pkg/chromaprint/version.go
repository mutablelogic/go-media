package chromaprint

import (
	"fmt"
	"io"

	// Packages
	"github.com/mutablelogic/go-media/sys/chromaprint"
)

func PrintVersion(w io.Writer) {
	fmt.Fprintf(w, "  %-10s %s\n", "chromaprint:", chromaprint.Version())
}
