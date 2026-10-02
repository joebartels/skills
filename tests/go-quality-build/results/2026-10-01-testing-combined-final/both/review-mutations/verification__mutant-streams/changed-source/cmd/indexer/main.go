package main

import (
	"context"
	"example.com/indexer"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const usage = "usage: indexer put DIR KEY TEXT | apply --dir DIR | watch --url URL --file PATH --interval DURATION"

func run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("%s", usage)
	}
	switch args[0] {
	case "put":
		if len(args) != 4 {
			return fmt.Errorf("%s", usage)
		}
		return indexer.Open(args[1]).Put(args[2], args[3])
	case "apply":
		flags := flag.NewFlagSet("apply", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		directory := flags.String("dir", "", "index directory")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if *directory == "" || flags.NArg() != 0 {
			return fmt.Errorf("%s", usage)
		}
		return indexer.Open(*directory).ApplyCSV(os.Stdin)
	case "watch":
		flags := flag.NewFlagSet("watch", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		endpoint := flags.String("url", "", "refresh URL")
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
			return fmt.Errorf("%s", usage)
		}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		transport := http.DefaultTransport.(*http.Transport).Clone()
		client := &http.Client{Transport: transport, Timeout: 30 * time.Second}
		return indexer.Serve(ctx, *interval, func(ctx context.Context) error {
			return indexer.Refresh(ctx, client, *endpoint, *path)
		}, func() error {
			client.CloseIdleConnections()
			return nil
		})
	default:
		return fmt.Errorf("%s", usage)
	}
}
func main() {
	fmt.Println("success")
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
