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

/*
 * Copyright (c) 2001 Fabrice Bellard
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 * THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */
