package indexer_test
import (
 "context"
 "fmt"
 "testing"
 "time"
 "example.com/indexer"
)
type reviewCauseError struct { details []string }
func (e reviewCauseError) Error() string { return fmt.Sprint(e.details) }
func TestReviewNonComparableCancellationCause(t *testing.T) {
 ctx, cancel := context.WithCancelCause(context.Background())
 defer cancel(nil)
 cause := reviewCauseError{details: []string{"shutdown"}}
 releases := 0
 defer func() {
  if p := recover(); p != nil { t.Errorf("Serve panicked for valid callback/cause error: %v; releases=%d", p, releases) }
 }()
 err := indexer.Serve(ctx, time.Second, func(got context.Context) error {
  cancel(cause)
  return fmt.Errorf("fetch stopped: %w", context.Cause(got))
 }, func() error { releases++; return nil })
 if err != nil || releases != 1 { t.Fatalf("Serve=%v releases=%d; want nil and one release", err, releases) }
}
