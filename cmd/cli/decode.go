package main

import (

	// Packages

	"encoding/json"
	"fmt"

	"github.com/mutablelogic/go-media"
)

type DecodeCmd struct {
	Path string `arg required help:"Media file" type:"path"`
}

func (cmd *DecodeCmd) Run(globals *Globals) error {
	reader, err := media.Open(cmd.Path, "")
	if err != nil {
		return err
	}
	defer reader.Close()

	data, _ := json.MarshalIndent(reader, "", "  ")
	fmt.Println(string(data))

	return nil
}
