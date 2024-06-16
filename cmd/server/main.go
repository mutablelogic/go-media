package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

var (
	port = flag.Int("port", 8080, "port to listen on")
)

func main() {
	mux := http.NewServeMux()

	// Output the metadata of a media file as JSON
	mux.HandleFunc("/metadata", handle_metadata)

	// Decode the audio or video data of a media file
	mux.HandleFunc("/decode", handle_decode)

	// Create the server, and listen
	server := http.Server{
		Addr:    fmt.Sprintf(":%v", *port),
		Handler: mux,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
