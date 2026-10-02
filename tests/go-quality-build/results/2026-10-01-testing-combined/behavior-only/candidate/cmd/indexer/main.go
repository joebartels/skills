package main

import (
	"context"
	"errors"
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

const usage = "usage: indexer put DIR KEY TEXT | apply --dir DIR | watch --url URL --file PATH --interval DURATION"

func run(args []string) error {
	return runContext(context.Background(), args, os.Stdin)
}

func runContext(ctx context.Context, args []string, input io.Reader) error {
	if len(args) == 0 {
		return errors.New(usage)
	}
	switch args[0] {
	case "put":
		if len(args) != 4 {
			return errors.New(usage)
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
			return errors.New(usage)
		}
		return indexer.Open(*dir).ApplyCSV(input)
	case "watch":
		flags := flag.NewFlagSet("watch", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		endpoint := flags.String("url", "", "records endpoint")
		path := flags.String("file", "", "destination path")
		interval := flags.Duration("interval", 0, "refresh interval")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		parsed, err := url.Parse(*endpoint)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" {
			return fmt.Errorf("watch requires an HTTP/HTTPS URL with a host")
		}
		if *path == "" || *interval <= 0 || flags.NArg() != 0 {
			return errors.New(usage)
		}
		transport := http.DefaultTransport.(*http.Transport).Clone()
		client := &http.Client{Transport: transport, Timeout: 30 * time.Second}
		return indexer.Serve(ctx, *interval, func(ctx context.Context) error {
			err := indexer.Refresh(ctx, client, *endpoint, *path)
			if ctx.Err() != nil && errors.Is(err, ctx.Err()) {
				return nil
			}
			return err
		}, func() error {
			client.CloseIdleConnections()
			return nil
		})
	default:
		return errors.New(usage)
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := runContext(ctx, os.Args[1:], os.Stdin); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
