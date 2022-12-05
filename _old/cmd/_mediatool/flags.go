package main

import "flag"

var (
	flags = []flag.Flag{
		{Name: "audio", Usage: "Filter audio files", DefValue: ""},
		{Name: "video", Usage: "Filter video files", DefValue: ""},
	}
)
