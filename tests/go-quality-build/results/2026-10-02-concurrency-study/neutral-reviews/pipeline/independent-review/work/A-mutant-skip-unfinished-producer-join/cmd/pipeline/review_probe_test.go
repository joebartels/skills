package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
)

func reviewEvent(t *testing.T, ch <-chan struct{}, label string) bool {
	t.Helper()
	select {
	case <-ch:
		return true
	case <-time.After(time.Second):
		t.Errorf("ASSERT missing event: %s", label)
		return false
	}
}

func TestReviewFIFO(t *testing.T) {
	for _, capacity := range []int{1, 3} {
		t.Run(fmt.Sprint(capacity), func(t *testing.T) {
			var got []int
			want := []int{5, 4, 3, 2, 1}
			err := runPipeline(context.Background(), capacity, func(ctx context.Context, out chan<- int) error {
				for _, value := range want {
					select { case out <- value: case <-ctx.Done(): return ctx.Err() }
				}
				return nil
			}, func(_ context.Context, in <-chan int) error {
				for value := range in { got = append(got, value) }
				return nil
			})
			if err != nil || !reflect.DeepEqual(got, want) { t.Errorf("ASSERT FIFO got=%v err=%v", got, err) }
		})
	}
}

func TestReviewEarlyStopJoin(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	blocked, stopped, release, producerDone, hostDone := make(chan struct{}), make(chan struct{}), make(chan struct{}), make(chan struct{}), make(chan struct{})
	result := make(chan error, 1)
	var once sync.Once
	t.Cleanup(func() { cancel(); once.Do(func() { close(release) }); reviewEvent(t, hostDone, "host cleanup"); reviewEvent(t, producerDone, "producer cleanup") })
	go func() {
		defer close(hostDone)
		result <- runPipeline(ctx, 1, func(ctx context.Context, out chan<- int) error {
			defer close(producerDone)
			out <- 1
			out <- 2
			close(blocked)
			select {
			case out <- 3: return errors.New("unexpected third send")
			case <-ctx.Done(): close(stopped); <-release; return ctx.Err()
			}
		}, func(_ context.Context, in <-chan int) error {
			<-in
			<-blocked
			return nil
		})
	}()
	if !reviewEvent(t, stopped, "early consumer must stop producer") { return }
	select { case <-hostDone: t.Error("ASSERT host returned before producer cleanup release"); default: }
	once.Do(func() { close(release) })
	if !reviewEvent(t, hostDone, "early host joins producer") { return }
	if err := <-result; err != nil { t.Errorf("ASSERT early completion error=%v", err) }
	select { case <-producerDone: default: t.Error("ASSERT producer not joined") }
}

func TestReviewFailureStopJoin(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ready, stopped, release, consumerDone, hostDone := make(chan struct{}), make(chan struct{}), make(chan struct{}), make(chan struct{}), make(chan struct{})
	result := make(chan error, 1)
	failure := errors.New("review producer failure")
	var once sync.Once
	t.Cleanup(func() { cancel(); once.Do(func() { close(release) }); reviewEvent(t, hostDone, "failure host cleanup"); reviewEvent(t, consumerDone, "consumer cleanup") })
	go func() {
		defer close(hostDone)
		result <- runPipeline(ctx, 1, func(context.Context, chan<- int) error { <-ready; return failure }, func(ctx context.Context, _ <-chan int) error {
			defer close(consumerDone)
			close(ready)
			<-ctx.Done()
			close(stopped)
			<-release
			return ctx.Err()
		})
	}()
	if !reviewEvent(t, stopped, "producer failure must stop consumer") { return }
	select { case <-hostDone: t.Error("ASSERT host returned before consumer cleanup release"); default: }
	once.Do(func() { close(release) })
	if !reviewEvent(t, hostDone, "failure host joins consumer") { return }
	if err := <-result; !errors.Is(err, failure) { t.Errorf("ASSERT independent producer error lost: %v", err) }
	select { case <-consumerDone: default: t.Error("ASSERT consumer not joined") }
}

func TestReviewIndependentExactBeforeStop(t *testing.T) {
	for _, producer := range []bool{true, false} {
		t.Run(fmt.Sprint(producer), func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			done := make(chan struct{})
			result := make(chan error, 1)
			t.Cleanup(func() { cancel(); reviewEvent(t, done, "independent exact cleanup") })
			independent := func(ctx context.Context) error {
				if ctx.Err() != nil { return errors.New("unexpected preexisting stop") }
				return context.Canceled
			}
			cooperative := func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() }
			go func() {
				defer close(done)
				result <- runPipeline(ctx, 1, func(ctx context.Context, _ chan<- int) error { if producer { return independent(ctx) }; return cooperative(ctx) }, func(ctx context.Context, _ <-chan int) error { if !producer { return independent(ctx) }; return cooperative(ctx) })
			}()
			if !reviewEvent(t, done, "independent exact completion") { return }
			if err := <-result; !errors.Is(err, context.Canceled) { t.Errorf("ASSERT independent exact before stop lost: %v", err) }
		})
	}
}

func TestReviewIndependentDeadlineAfterStop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	result := make(chan error, 1)
	t.Cleanup(func() { cancel(); reviewEvent(t, done, "independent deadline cleanup") })
	go func() {
		defer close(done)
		result <- runPipeline(ctx, 1, func(ctx context.Context, _ chan<- int) error {
			<-ctx.Done()
			if ctx.Err() != context.Canceled { return errors.New("unexpected stop class") }
			// A private operation independently expires during cleanup. The pipeline
			// stop is Canceled, so DeadlineExceeded cannot be its stop error.
			return context.DeadlineExceeded
		}, func(context.Context, <-chan int) error { return nil })
	}()
	if !reviewEvent(t, done, "deadline failure completion") { return }
	if err := <-result; !errors.Is(err, context.DeadlineExceeded) { t.Errorf("ASSERT independent deadline during cleanup lost: %v", err) }
}

func TestReviewWrappedFailures(t *testing.T) {
	producerFailure := errors.Join(context.Canceled, errors.New("producer operation"))
	consumerFailure := fmt.Errorf("cleanup: %w", errors.Join(context.Canceled, errors.New("consumer operation")))
	err := runPipeline(context.Background(), 1, func(context.Context, chan<- int) error { return producerFailure }, func(ctx context.Context, _ <-chan int) error { <-ctx.Done(); return consumerFailure })
	if !errors.Is(err, producerFailure) || !errors.Is(err, consumerFailure) { t.Errorf("ASSERT wrapped independent failures lost: %v", err) }
}

type reviewFailWriter struct { accepted bytes.Buffer; calls int; failure error }
func (w *reviewFailWriter) Write(p []byte) (int, error) { w.calls++; if w.calls == 2 { return 0, w.failure }; return w.accepted.Write(p) }
func TestReviewWriterFailurePreservesAcceptedOutput(t *testing.T) {
	w := &reviewFailWriter{failure: errors.New("writer failed")}
	var diagnostic bytes.Buffer
	code := command(context.Background(), nil, strings.NewReader("9\n8\n7\n"), w, &diagnostic)
	if code != 2 || w.accepted.String() != "9\n" || !strings.Contains(diagnostic.String(), "writer failed") { t.Errorf("ASSERT writer failure code=%d accepted=%q diagnostic=%q", code, w.accepted.String(), diagnostic.String()) }
}
