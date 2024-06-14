package main

import (
	"flag"
	"log"
)

func main() {
	in := flag.String("in", "", "input file")
	out := flag.String("out", "", "output file")
	flag.Parse()

	// Check in and out
	if *in == "" || *out == "" {
		log.Fatal("-in and -out files must be specified")
	}
}
