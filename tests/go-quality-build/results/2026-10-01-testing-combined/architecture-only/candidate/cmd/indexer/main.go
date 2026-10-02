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

	"example.com/indexer"
)

func run(args []string) error {
	return runContext(context.Background(), args, os.Stdin)
}

func runContext(ctx context.Context, args []string, stdin io.Reader) error {
	if len(args) == 0 {
		return usage()
	}
	switch args[0] {
	case "put":
		if len(args) != 4 {
			return usage()
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
		flags := flag.NewFlagSet("watch", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		endpoint := flags.String("url", "", "records URL")
		path := flags.String("file", "", "destination file")
		interval := flags.Duration("interval", 0, "refresh interval")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		parsed, err := url.Parse(*endpoint)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" {
			return fmt.Errorf("watch URL must be HTTP or HTTPS with a host")
		}
		if *path == "" || *interval <= 0 || flags.NArg() != 0 {
			return fmt.Errorf("usage: indexer watch --url URL --file PATH --interval DURATION (positive)")
		}
		transport := &http.Transport{Proxy: http.ProxyFromEnvironment}
		client := &http.Client{Transport: transport}
		return indexer.Serve(ctx, *interval, func(ctx context.Context) error {
			return indexer.Refresh(ctx, client, *endpoint, *path)
		}, func() error {
			transport.CloseIdleConnections()
			return nil
		})
	default:
		return usage()
	}
}

func usage() error {
	return fmt.Errorf("usage: indexer put DIR KEY TEXT | apply --dir DIR | watch --url URL --file PATH --interval DURATION")
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := runContext(ctx, os.Args[1:], os.Stdin)
	stop()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
