package main

import (
	"encoding/json"
	"fmt"
	"os"

	"example.com/showcfg"
)

func main() {
	config, err := showcfg.LoadFromEnv()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	data, marshalErr := json.Marshal(config)
	if marshalErr != nil { panic(marshalErr) }
	if _, err := os.Stdout.Write(data); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
