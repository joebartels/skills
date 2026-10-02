package main

import (
	"context"
	"example.com/indexer"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage")
	}
	switch args[0] {
	case "put":
		if len(args) != 4 {
			return fmt.Errorf("usage")
		}
		return indexer.Open(args[1]).Put(args[2], args[3])
	case "apply":
		fs := flag.NewFlagSet("apply", flag.ContinueOnError)
		dir := fs.String("dir", "", "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if *dir == "" || fs.NArg() != 0 {
			return fmt.Errorf("dir required")
		}
		return indexer.Open(*dir).ApplyCSV(os.Stdin)
	case "watch":
		fs := flag.NewFlagSet("watch", flag.ContinueOnError)
		endpoint := fs.String("url", "", "")
		path := fs.String("file", "", "")
		interval := fs.Duration("interval", time.Second, "")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		u, err := url.Parse(*endpoint)
		if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || *path == "" || *interval <= 0 || fs.NArg() != 0 {
			return fmt.Errorf("invalid watch arguments")
		}
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()
		client := &http.Client{}
		return indexer.Serve(ctx, *interval, func(c context.Context) error { return indexer.Refresh(c, client, *endpoint, *path) }, func() error { client.CloseIdleConnections(); return nil })
	default:
		return fmt.Errorf("unknown command")
	}
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
