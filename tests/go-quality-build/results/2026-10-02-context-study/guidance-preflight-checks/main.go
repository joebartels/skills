package main

import (
    "context"
    "errors"
    "fmt"
    "time"
)

func finalizeAccepted(ctx context.Context, budget time.Duration, finish func(context.Context) error) error {
	finalCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), budget)
	defer cancel()
	return finish(finalCtx)
}

type metadataKey struct{}
type sliceCause []string
func (s sliceCause) Error() string { return "caller stopped" }

func require(ok bool, message string) {
    if !ok { panic(message) }
}

func liveTransferredScope(parent context.Context) (context.Context, context.CancelFunc) {
    child, cancel := context.WithCancel(parent)
    return child, cancel
}
func prematureCreatorCancel(parent context.Context) (context.Context, context.CancelFunc) {
    child, cancel := context.WithCancel(parent)
    defer cancel()
    return child, cancel
}

func main() {
    parent, cancelParent := context.WithCancelCause(context.WithValue(context.Background(), metadataKey{}, "request-id"))
    cause := sliceCause{"legal", "non-comparable"}
    cancelParent(cause)
    classification := parent.Err()
    gotCause := context.Cause(parent)
    independent := errors.New("operation failed")
    combined := errors.Join(independent, classification, gotCause)
    var observed sliceCause
    require(errors.Is(combined, independent), "independent error lost")
    require(errors.Is(combined, context.Canceled), "classification lost")
    require(errors.As(combined, &observed) && len(observed)==2, "non-comparable cause lost")

    detached := context.WithoutCancel(parent)
    _, deadlinePresent := detached.Deadline()
    require(detached.Done()==nil && detached.Err()==nil && context.Cause(detached)==nil && !deadlinePresent, "detachment retained stop signals")
    require(detached.Value(metadataKey{})=="request-id", "request metadata lost")
    finalFailure := errors.New("finalization failed")
    var finalScope context.Context
    finalErr := finalizeAccepted(parent, 10*time.Second, func(ctx context.Context) error {
        finalScope = ctx
        _, bounded := ctx.Deadline()
        require(ctx.Err()==nil && bounded, "finalization scope not live and bounded")
        require(ctx.Value(metadataKey{})=="request-id", "finalization metadata lost")
        return finalFailure
    })
    require(errors.Is(finalErr, finalFailure), "finalization error lost")
    require(finalScope.Err()==context.Canceled, "finalization cancel not released on return")

    earlier, cancelEarlier := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancelEarlier()
    total, cancelTotal := context.WithTimeout(earlier, time.Minute)
    defer cancelTotal()
    stage1, cancel1 := context.WithTimeout(total, time.Minute)
    defer cancel1()
    stage2, cancel2 := context.WithTimeout(total, time.Minute)
    defer cancel2()
    parentDeadline, _ := earlier.Deadline()
    totalDeadline, _ := total.Deadline()
    stage1Deadline, _ := stage1.Deadline()
    stage2Deadline, _ := stage2.Deadline()
    require(parentDeadline.Equal(totalDeadline) && totalDeadline.Equal(stage1Deadline) && stage1Deadline.Equal(stage2Deadline), "shared/earlier budget lost")

    transferred, transferredCancel := liveTransferredScope(context.Background())
    require(transferred.Err()==nil, "explicit transferred scope stopped before use")
    transferredCancel()
    require(transferred.Err()==context.Canceled, "transferred owner did not cancel")
    premature, prematureCancel := prematureCreatorCancel(context.Background())
    defer prematureCancel()
    require(premature.Err()==context.Canceled, "creator defer counterexample not reproduced")

    callbackCtx, callbackCancel := context.WithCancel(context.Background())
    started, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
    stop := context.AfterFunc(callbackCtx, func() {
        close(started)
        <-release
        close(done)
    })
    callbackCancel()
    select {
    case <-started:
    case <-time.After(time.Second): panic("callback did not start")
    }
    require(!stop(), "already-started callback unexpectedly prevented")
    select {
    case <-done: panic("callback completed before release")
    default:
    }
    close(release)
    select {
    case <-done:
    case <-time.After(time.Second): panic("callback did not complete")
    }
    fmt.Println("PASS: exact finalization snippet; non-comparable cause inspection; shared/earlier budgets; cancel-transfer counterexample; AfterFunc stop/completion distinction")
}
