package main

import (
	"context"
	"example.com/commerce/internal/billing"
	"example.com/commerce/internal/gateway"
	"example.com/commerce/internal/invoice"
	"fmt"
	"net/http"
	"os"
	"time"
)

func run() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("usage: settle FILE")
	}
	input, err := os.Open(os.Args[1])
	if err != nil {
		return err
	}
	defer input.Close()
	ledger, err := os.OpenFile(os.Getenv("LEDGER"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer ledger.Close()
	service := billing.Service{Ledger: ledger, Gateway: gateway.HTTP{Client: &http.Client{Timeout: 5 * time.Second}, URL: os.Getenv("GATEWAY_URL")}}
	return invoice.Read(input, func(row invoice.Row) error {
		return service.Charge(context.Background(), row.ID, row.Cents)
	})
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
