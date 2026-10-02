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

	"example.com/indexer"
)

func run(args []string) error { return runContext(context.Background(), args, os.Stdin) }

func runContext(ctx context.Context, args []string, input io.Reader) error {
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
		if flags.NArg() != 0 || *dir == "" {
			return fmt.Errorf("usage: indexer apply --dir DIR")
		}
		return indexer.Open(*dir).ApplyCSV(input)
	case "watch":
		flags := flag.NewFlagSet("watch", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		endpoint := flags.String("url", "", "HTTP or HTTPS source")
		path := flags.String("file", "", "snapshot path")
		interval := flags.Duration("interval", 0, "delay between completed refreshes")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		source, err := url.Parse(*endpoint)
		if err != nil || (source.Scheme != "http" && source.Scheme != "https") || source.Hostname() == "" || *path == "" || *interval <= 0 || flags.NArg() != 0 {
			return fmt.Errorf("usage: indexer watch --url HTTP_URL --file PATH --interval POSITIVE_DURATION")
		}
		client := &http.Client{Transport: http.DefaultTransport.(*http.Transport).Clone()}
		return indexer.Serve(ctx, *interval,
			func(ctx context.Context) error {
				err := indexer.Refresh(ctx, client, *endpoint, *path)
				if ctx.Err() != nil && errors.Is(err, ctx.Err()) {
					return nil
				}
				return err
			},
			func() error { client.CloseIdleConnections(); return nil },
		)
	default:
		return fmt.Errorf("unknown command: %s", args[0])
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := runContext(ctx, os.Args[1:], os.Stdin); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
