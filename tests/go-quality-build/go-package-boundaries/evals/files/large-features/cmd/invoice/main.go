package main

import (
	"context"
	"example.com/commerce/internal/billing"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

func run() error {
	if len(os.Args) != 3 {
		return fmt.Errorf("usage: invoice ID CENTS")
	}
	n, err := strconv.Atoi(os.Args[2])
	if err != nil {
		return err
	}
	f, err := os.OpenFile(os.Getenv("LEDGER"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	s := billing.Service{Ledger: f, Client: &http.Client{Timeout: 5 * time.Second}, GatewayURL: os.Getenv("GATEWAY_URL")}
	return s.Charge(context.Background(), os.Args[1], n)
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
