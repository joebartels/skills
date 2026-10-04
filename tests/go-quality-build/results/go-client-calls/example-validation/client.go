package remote

import (
    "context"
    "errors"
    "fmt"
    "io"
    "net/http"
)

func FetchOnce(ctx context.Context, client *http.Client, endpoint string, maxBytes int64) ([]byte, error) {
    if maxBytes < 0 || maxBytes == 1<<63-1 {
        return nil, errors.New("invalid response limit")
    }
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
    if err != nil {
        return nil, err
    }
    resp, err := client.Do(req)
    if err != nil {
        return nil, err // An error response's redirect body is already closed.
    }
    data, readErr := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
    closeErr := resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return nil, errors.Join(fmt.Errorf("HTTP %d", resp.StatusCode), readErr, closeErr)
    }
    if int64(len(data)) > maxBytes {
        return nil, errors.Join(errors.New("response exceeds limit"), readErr, closeErr)
    }
    if err := errors.Join(readErr, closeErr); err != nil {
        return nil, err
    }
    return data, nil
}
