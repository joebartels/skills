package indexer
import (
    "context"
    "errors"
    "time"
)
var ErrInvalidInterval = errors.New("invalid interval")
func Serve(ctx context.Context, interval time.Duration, refresh func(context.Context) error, release func() error) (err error) {
    if interval <= 0 { return ErrInvalidInterval }
    released := false
    defer func() { if !released { err = errors.Join(err, release()) } }()
    for {
        if ctx.Err() != nil { return nil }
        result := make(chan error, 1)
        go func() { result <- refresh(ctx) }()
        select {
        case callbackErr := <-result:
            if callbackErr != nil { return callbackErr }
        case <-ctx.Done():
            releaseErr := release()
            released = true
            return errors.Join(<-result, releaseErr)
        }
        timer := time.NewTimer(interval)
        select {
        case <-ctx.Done(): timer.Stop(); return nil
        case <-timer.C:
        }
    }
}
