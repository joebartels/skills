package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"example.com/indexer"
)

func run(args []string) error {
	return runInput(args, os.Stdin)
}

func runInput(args []string, input io.Reader) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: indexer put DIR KEY TEXT | apply --dir DIR | watch --url URL --file PATH --interval DURATION")
	}
	switch args[0] {
	case "put":
		// Preserve positional arguments, including TEXT beginning with '-'.
		if len(args) != 4 {
			return fmt.Errorf("usage: indexer put DIR KEY TEXT")
		}
		return indexer.Open(args[1]).Put(args[2], args[3])
	case "apply":
		flags := flag.NewFlagSet("apply", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		dir := flags.String("dir", "", "index directory")
		if err := flags.Parse(args[1:]); err != nil {
			return fmt.Errorf("apply arguments: %w", err)
		}
		if *dir == "" || flags.NArg() != 0 {
			return fmt.Errorf("usage: indexer apply --dir DIR")
		}
		return indexer.Open(*dir).ApplyCSV(input)
	case "watch":
		flags := flag.NewFlagSet("watch", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		endpoint := flags.String("url", "", "snapshot URL")
		path := flags.String("file", "", "snapshot path")
		interval := flags.Duration("interval", 0, "refresh interval")
		if err := flags.Parse(args[1:]); err != nil {
			return fmt.Errorf("watch arguments: %w", err)
		}
		parsed, err := url.Parse(*endpoint)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" || *path == "" || *interval <= 0 || flags.NArg() != 0 {
			return fmt.Errorf("usage: indexer watch --url HTTP_URL --file PATH --interval POSITIVE_DURATION")
		}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		client := &http.Client{Timeout: 30 * time.Second}
		return indexer.Serve(ctx, *interval, func(ctx context.Context) error {
			return indexer.Refresh(ctx, client, *endpoint, *path)
		}, func() error {
			client.CloseIdleConnections()
			return nil
		})
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
