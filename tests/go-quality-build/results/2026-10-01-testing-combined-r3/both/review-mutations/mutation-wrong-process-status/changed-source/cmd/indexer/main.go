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
	"strings"
	"syscall"
	"time"

	"example.com/indexer"
)

func run(args []string) error {
	return runContext(context.Background(), args, os.Stdin)
}

func runContext(ctx context.Context, args []string, stdin io.Reader) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: indexer put DIR KEY TEXT | apply --dir DIR | watch --url URL --file PATH --interval DURATION")
	}
	switch args[0] {
	case "put":
		// Preserve the original positional grammar, including leading dashes.
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
		return indexer.Open(*dir).ApplyCSV(stdin)
	case "watch":
		return watch(ctx, args[1:])
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func watch(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("watch", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	endpoint := flags.String("url", "", "HTTP endpoint")
	path := flags.String("file", "", "snapshot path")
	interval := flags.Duration("interval", 0, "time between completed refreshes")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 || *path == "" || *interval <= 0 {
		return fmt.Errorf("usage: indexer watch --url URL --file PATH --interval DURATION (positive interval required)")
	}
	parsed, err := url.Parse(*endpoint)
	if err != nil {
		return fmt.Errorf("invalid watch URL: %w", err)
	}
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	if (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" {
		return fmt.Errorf("watch URL must use HTTP or HTTPS with a host")
	}
	client := &http.Client{Timeout: 30 * time.Second}
	return indexer.Serve(ctx, *interval, func(ctx context.Context) error {
		return indexer.Refresh(ctx, client, parsed.String(), *path)
	}, func() error {
		client.CloseIdleConnections()
		return nil
	})
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := runContext(ctx, os.Args[1:], os.Stdin); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
