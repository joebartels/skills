package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func waitPipeline(t *testing.T, ch <-chan struct{}, what string) bool {
	t.Helper()
	select {
	case <-ch:
		return true
	case <-time.After(time.Second):
		t.Errorf("missing event: %s", what)
		return false
	}
}
func TestControllerEarlyConsumerAndProducerFailure(t *testing.T) {
	t.Run("early consumer joins blocked producer", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		attemptThird := make(chan struct{})
		stopped := make(chan struct{})
		release := make(chan struct{})
		producerDone := make(chan struct{})
		consumerDone := make(chan struct{})
		hostDone := make(chan struct{})
		result := make(chan error, 1)
		var once sync.Once
		t.Cleanup(func() {
			cancel()
			once.Do(func() { close(release) })
			waitPipeline(t, hostDone, "pipeline cleanup")
			waitPipeline(t, producerDone, "producer cleanup")
		})
		go func() {
			defer close(hostDone)
			result <- runPipeline(ctx, 1, func(ctx context.Context, out chan<- int) error {
				defer close(producerDone)
				for _, v := range []int{1, 2} {
					select {
					case out <- v:
					case <-ctx.Done():
						return ctx.Err()
					}
				}
				close(attemptThird)
				select {
				case out <- 3:
					return errors.New("unexpected third acceptance")
				case <-ctx.Done():
					close(stopped)
					<-release
					return ctx.Err()
				}
			}, func(ctx context.Context, in <-chan int) error {
				defer close(consumerDone)
				select {
				case value := <-in:
					if value != 1 {
						return fmt.Errorf("first=%d", value)
					}
				case <-ctx.Done():
					return ctx.Err()
				}
				<-attemptThird
				return nil
			})
		}()
		if !waitPipeline(t, consumerDone, "consumer completion") || !waitPipeline(t, stopped, "blocked producer requested to stop") {
			return
		}
		select {
		case <-hostDone:
			t.Error("returned before producer cleanup joined")
		default:
		}
		once.Do(func() { close(release) })
		if !waitPipeline(t, hostDone, "joined early completion") {
			return
		}
		if err := <-result; err != nil {
			t.Errorf("normal early completion failed: %v", err)
		}
		select {
		case <-producerDone:
		default:
			t.Error("producer not joined")
		}
	})
	t.Run("producer failure stops blocked receive", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		failure := errors.New("independent producer failure")
		ready := make(chan struct{})
		consumerDone := make(chan struct{})
		hostDone := make(chan struct{})
		result := make(chan error, 1)
		var accepted int
		t.Cleanup(func() { cancel(); waitPipeline(t, hostDone, "failed pipeline cleanup") })
		go func() {
			defer close(hostDone)
			result <- runPipeline(ctx, 1, func(ctx context.Context, out chan<- int) error {
				select {
				case out <- 7:
				case <-ctx.Done():
					return ctx.Err()
				}
				<-ready
				return failure
			}, func(ctx context.Context, in <-chan int) error {
				defer close(consumerDone)
				select {
				case value := <-in:
					accepted = value
				case <-ctx.Done():
					return ctx.Err()
				}
				close(ready)
				<-ctx.Done()
				return ctx.Err()
			})
		}()
		if !waitPipeline(t, hostDone, "producer failure joins consumer") {
			return
		}
		if err := <-result; !errors.Is(err, failure) {
			t.Errorf("independent failure lost: %v", err)
		}
		if accepted != 7 {
			t.Errorf("accepted=%d", accepted)
		}
		select {
		case <-consumerDone:
		default:
			t.Error("consumer not joined")
		}
	})
}
func TestControllerActualPipelineCommand(t *testing.T) {
	tool := os.Getenv("GOQUALITY_GO")
	if tool == "" {
		tool = "go"
	}
	binary := filepath.Join(t.TempDir(), "pipeline")
	build := exec.Command(tool, "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v %s", err, output)
	}
	for _, c := range []struct {
		name, input, want string
		code              int
	}{{"early prefix", "1\n2\n3\n4\n5\n", "1\n2\n", 0}, {"malformed first", "bad\n1\n", "", 2}} {
		t.Run(c.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, binary, "--take", "2")
			cmd.Stdin = strings.NewReader(c.input)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err := cmd.Run()
			code := 0
			if err != nil {
				var exit *exec.ExitError
				if !errors.As(err, &exit) {
					t.Fatalf("process: %v", err)
				}
				code = exit.ExitCode()
			}
			if ctx.Err() != nil {
				t.Fatalf("pipeline did not terminate: stdout=%q stderr=%q", stdout.String(), stderr.String())
			}
			if code != c.code || stdout.String() != c.want {
				t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
			}
			if c.code == 0 && stderr.Len() != 0 {
				t.Errorf("normal diagnostic=%q", stderr.String())
			}
			if c.code == 2 && !strings.Contains(stderr.String(), "integer input") {
				t.Errorf("missing failure diagnostic=%q", stderr.String())
			}
		})
	}
}

func TestControllerIndependentCancellationClassBeforeStop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ready := make(chan struct{})
	done := make(chan struct{})
	result := make(chan error, 1)
	t.Cleanup(func() { cancel(); waitPipeline(t, done, "independent cancellation-class pipeline cleanup") })
	go func() {
		defer close(done)
		result <- runPipeline(ctx, 1, func(ctx context.Context, out chan<- int) error {
			select {
			case out <- 7:
			case <-ctx.Done():
				return ctx.Err()
			}
			<-ready
			return context.Canceled
		}, func(ctx context.Context, in <-chan int) error {
			select {
			case <-in:
				close(ready)
			case <-ctx.Done():
				return ctx.Err()
			}
			<-ctx.Done()
			return ctx.Err()
		})
	}()
	if !waitPipeline(t, done, "independent cancellation class retained after coordinated stopping") {
		return
	}
	if err := <-result; !errors.Is(err, context.Canceled) {
		t.Errorf("independent cancellation-class failure lost: %v", err)
	}
}
