/*
An example of fingerprinting audio and recognizing the any music tracks within the audio.
*/
package main

import (
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	// Packages
	chromaprint "github.com/mutablelogic/go-media/pkg/chromaprint"
	config "github.com/mutablelogic/go-media/pkg/config"
)

var (
	flagVersion  = flag.Bool("version", false, "Print version information")
	flagRate     = flag.Int("rate", 22050, "Sample rate")
	flagChannels = flag.Int("channels", 1, "Number of channels")
	flagKey      = flag.String("key", "${CHROMAPRINT_KEY}", "AcoustID API key")
	flagLength   = flag.Duration("length", 2*time.Minute, "Restrict the duration of the processed input audio")
	flagDuration = flag.Duration("duration", 0, "The actual duration of the audio file")
)

const (
	bufsize = 1024 * 64 // Number of bytes to read at a time
)

func main() {
	flag.Parse()

	// Check for version
	if *flagVersion {
		config.PrintVersion(flag.CommandLine.Output())
		chromaprint.PrintVersion(flag.CommandLine.Output())
		os.Exit(0)
	}

	// Check arguments
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(-1)
	}

	// Open file
	info, err := os.Stat(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-2)
	}
	if info.Size() == 0 {
		fmt.Fprintln(os.Stderr, "file is empty")
		os.Exit(-2)
	}

	// Open the file
	r, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-2)
	}
	defer r.Close()

	// Create fingerprinter, write samples
	size := int(info.Size() >> 1)
	fingerprint := chromaprint.New(*flagRate, *flagChannels, *flagLength)
	samples := make([]int16, bufsize)
	for {
		// Adjust buffer size
		sz := MinInt(bufsize, size)
		size -= sz
		if sz == 0 {
			break
		}

		// Read samples
		if err := binary.Read(r, binary.LittleEndian, samples[:sz]); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			fmt.Fprintln(os.Stderr, "Unexpected error:", err)
			os.Exit(-2)
		}

		// Write samples
		if _, err := fingerprint.Write(samples); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(-2)
		}
	}

	// Get fingerprint
	str, err := fingerprint.Finish()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-2)
	}

	// Get duration
	duration := fingerprint.Duration()
	if *flagDuration != 0 {
		duration = *flagDuration
	}

	// Create client, make matches
	client := chromaprint.NewClient(*flagKey)
	if matches, err := client.Lookup(str, duration, chromaprint.META_ALL); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(-2)
	} else if len(matches) == 0 {
		fmt.Fprintln(os.Stderr, "No matches found")
	} else {
		for _, match := range matches {
			fmt.Println(match)
		}
	}
}

func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
