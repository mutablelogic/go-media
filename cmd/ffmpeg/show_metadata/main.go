package main

import (
	"flag"
	"log"

	// Packages
	ff "github.com/mutablelogic/go-media/sys/ffmpeg61"
)

func main() {
	in := flag.String("in", "", "input file")
	flag.Parse()

	// Check out and size
	if *in == "" {
		log.Fatal("-in argument must be specified")
	}

	// Open input file
	input, err := ff.AVFormat_open_url(*in, nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ff.AVFormat_close_input(input)

	if err := ff.AVFormat_find_stream_info(input, nil); err != nil {
		log.Fatal(err)
	}

	for _, tag := range ff.AVUtil_dict_entries(input.Metadata()) {
		log.Println(tag.Key(), "=>", tag.Value())
	}
}
