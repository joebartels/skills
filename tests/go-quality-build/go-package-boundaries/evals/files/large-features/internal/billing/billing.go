package billing

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Service is used serially by the host. Invoices are append-only lines: ID,CENTS.
type Service struct {
	Ledger     io.Writer
	Client     *http.Client
	GatewayURL string
}

func (s *Service) Charge(ctx context.Context, id string, cents int) error {
	if id == "" || strings.ContainsAny(id, ",\r\n") || cents <= 0 {
		return fmt.Errorf("invalid invoice")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.GatewayURL, strings.NewReader(id+","+strconv.Itoa(cents)))
	if err != nil {
		return err
	}
	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 204 {
		return fmt.Errorf("gateway status %d", resp.StatusCode)
	}
	_, err = fmt.Fprintf(s.Ledger, "%s,%d\n", id, cents)
	return err
}

// Export copies the existing ledger bytes to out.
func Export(path string, out io.Writer) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(out, f)
	return err
}
