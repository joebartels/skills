package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"time"

	"example.com/finalize"
)

func main() {
	receipt := flag.String("receipt", "", "accepted-count receipt path")
	budget := flag.Duration("finalization-budget", 2*time.Second, "required finalization budget")
	flag.Parse()
	if *receipt == "" || *budget <= 0 {
		fmt.Fprintln(os.Stderr, "receipt and positive budget required")
		os.Exit(2)
	}
	var jobs []int
	for _, arg := range flag.Args() {
		n, err := strconv.Atoi(arg)
		if err != nil || n < 0 {
			fmt.Fprintln(os.Stderr, "invalid job")
			os.Exit(2)
		}
		jobs = append(jobs, n)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	_, err := finalize.Run(ctx, jobs, func(ctx context.Context, job int) error {
		if job == 0 {
			fmt.Fprintln(os.Stderr, "waiting")
			<-ctx.Done()
			return ctx.Err()
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		_, err := fmt.Fprintln(os.Stdout, "accepted", job)
		return err
	}, func(ctx context.Context, accepted int) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		return os.WriteFile(*receipt, []byte(fmt.Sprintf("accepted=%d\n", accepted)), 0600)
	}, *budget)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}
