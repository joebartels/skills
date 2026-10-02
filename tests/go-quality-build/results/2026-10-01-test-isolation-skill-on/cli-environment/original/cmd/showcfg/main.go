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
	if err := json.NewEncoder(os.Stdout).Encode(config); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
