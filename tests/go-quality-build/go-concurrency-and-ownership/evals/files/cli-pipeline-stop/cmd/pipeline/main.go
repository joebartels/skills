package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func command(ctx context.Context, args []string, input io.Reader, output, diagnostic io.Writer) int {
	flags := flag.NewFlagSet("pipeline", flag.ContinueOnError)
	flags.SetOutput(diagnostic)
	take := flags.Int("take", 1000000, "maximum accepted integers")
	if err := flags.Parse(args); err != nil || *take <= 0 || flags.NArg() != 0 {
		return 2
	}
	err := runPipeline(ctx, 1, func(ctx context.Context, out chan<- int) error {
		scanner := bufio.NewScanner(input)
		for scanner.Scan() {
			value, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
			if err != nil {
				return fmt.Errorf("integer input: %w", err)
			}
			select {
			case out <- value:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return scanner.Err()
	}, func(ctx context.Context, in <-chan int) error {
		for accepted := 0; accepted < *take; accepted++ {
			select {
			case value, ok := <-in:
				if !ok {
					return nil
				}
				if _, err := fmt.Fprintln(output, value); err != nil {
					return err
				}
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		return nil
	})
	if err != nil {
		fmt.Fprintln(diagnostic, err)
		return 2
	}
	return 0
}
func main() { os.Exit(command(context.Background(), os.Args[1:], os.Stdin, os.Stdout, os.Stderr)) }
