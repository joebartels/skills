package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"example.com/dispatch"
)

func run(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: replay FILE")
	}
	file, err := os.Open(args[0])
	if err != nil {
		return err
	}
	defer file.Close()

	s := &dispatch.Service{
		Dir:       os.Getenv("RECORD_DIR"),
		NotifyURL: os.Getenv("NOTIFY_URL"),
		Client:    &http.Client{Timeout: 5 * time.Second},
	}
	return s.Replay(context.Background(), file)
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
