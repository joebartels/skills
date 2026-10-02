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
	if len(args) == 0 {
		return fmt.Errorf("usage: indexer put DIR KEY TEXT | apply --dir DIR | watch --url URL --file PATH --interval DURATION")
	}
	switch args[0] {
	case "put":
		if len(args) != 4 {
			return fmt.Errorf("usage: indexer put DIR KEY TEXT")
		}
		return indexer.Open(args[1]).Put(args[2], args[3])
	case "apply":
		flags := flag.NewFlagSet("apply", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		dir := flags.String("dir", "", "index directory")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if *dir == "" || flags.NArg() != 0 {
			return fmt.Errorf("usage: indexer apply --dir DIR")
		}
		return indexer.Open(*dir).ApplyCSV(os.Stdin)
	case "watch":
		return watch(args[1:])
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func watch(args []string) error {
	flags := flag.NewFlagSet("watch", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	endpoint := flags.String("url", "", "record endpoint")
	path := flags.String("file", "", "snapshot destination")
	interval := flags.Duration("interval", 0, "time between completed refreshes")
	if err := flags.Parse(args); err != nil {
		return err
	}
	u, err := url.Parse(*endpoint)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || *path == "" || *interval <= 0 || flags.NArg() != 0 {
		return fmt.Errorf("usage: indexer watch --url URL --file PATH --interval DURATION (HTTP/HTTPS URL and positive interval required)")
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
}
func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
