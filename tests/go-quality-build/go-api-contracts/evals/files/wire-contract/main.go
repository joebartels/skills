package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type job struct {
	ID    string `json:"id"`
	State string `json:"state"`
}
type response struct {
	Jobs []job `json:"jobs"`
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: joblist FILE")
		return 2
	}
	data, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	var jobs []job
	if err := json.Unmarshal(data, &jobs); err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	var result []job
	for _, item := range jobs {
		result = append(result, item)
	}
	if err := json.NewEncoder(stdout).Encode(response{Jobs: result}); err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	return 0
}

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }
