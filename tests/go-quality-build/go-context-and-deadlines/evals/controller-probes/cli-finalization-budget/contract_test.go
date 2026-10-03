package finalize

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestFinalizationAfterCancel(t *testing.T) {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	cause := errors.New("processing interrupted")
	failure := errors.New("receipt persistence failed")
	entered, release := make(chan struct{}), make(chan struct{})
	type outcome struct {
		n   int
		err error
	}
	done := make(chan outcome, 1)
	go func() {
		n, err := Run(ctx, []int{1, 2, 3}, func(_ context.Context, job int) error {
			if job == 2 {
				cancel(cause)
				return context.Canceled
			}
			return nil
		}, func(finalCtx context.Context, n int) error {
			close(entered)
			if n != 1 {
				return errors.New("wrong accepted count")
			}
			deadline, ok := finalCtx.Deadline()
			if finalCtx.Err() != nil || !ok || deadline.After(time.Now().Add(time.Second)) {
				return errors.New("finalization scope not live and bounded")
			}
			<-release
			return failure
		}, time.Second)
		done <- outcome{n, err}
	}()
	select {
	case <-entered:
	case got := <-done:
		close(release)
		t.Fatalf("finalization not entered: %+v", got)
	case <-time.After(2 * time.Second):
		cancel(nil)
		close(release)
		t.Fatal("finalization did not start")
	}
	select {
	case got := <-done:
		close(release)
		t.Fatalf("returned before finalization completion: %+v", got)
	default:
	}
	close(release)
	select {
	case got := <-done:
		if got.n != 1 || !errors.Is(got.err, context.Canceled) || !errors.Is(got.err, cause) || !errors.Is(got.err, failure) {
			t.Fatalf("accepted=%d err=%v", got.n, got.err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("finalization did not complete")
	}
}

func TestFinalizationDeadline(t *testing.T) {
	n, err := Run(context.Background(), []int{1}, func(context.Context, int) error { return nil }, func(ctx context.Context, _ int) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
			return errors.New("unbounded finalization")
		}
	}, 30*time.Millisecond)
	if n != 1 || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("accepted=%d err=%v", n, err)
	}
}

func TestUnacceptedCanceledWork(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls, finalized := 0, 0
	n, err := Run(ctx, []int{1}, func(context.Context, int) error { calls++; return nil }, func(context.Context, int) error { finalized++; return nil }, time.Second)
	if calls != 0 || finalized != 0 || n != 0 || !errors.Is(err, context.Canceled) {
		t.Fatalf("calls=%d finalized=%d accepted=%d err=%v", calls, finalized, n, err)
	}
}

func buildCommand(t *testing.T) string {
	t.Helper()
	goExe := os.Getenv("GOQUALITY_GO")
	if goExe == "" {
		goExe = "go"
	}
	root := t.TempDir()
	binary := filepath.Join(root, "finalize")
	cmd := exec.Command(goExe, "build", "-o", binary, "./cmd/finalize")
	cmd.Env = os.Environ()
	if os.Getenv("GOCACHE") == "" {
		cmd.Env = append(cmd.Env, "GOCACHE="+filepath.Join(root, "cache"))
	}
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}
	return binary
}

func TestFinalizationAfterCancelProcess(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fixture interruption host uses os.Interrupt")
	}
	binary := buildCommand(t)
	receipt := filepath.Join(t.TempDir(), "receipt")
	guard, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(guard, binary, "--receipt", receipt, "1", "0")
	cmd.Env = []string{"PATH=" + os.Getenv("PATH")}
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	observed := make(chan bool, 1)
	go func() {
		scanner := bufio.NewScanner(stderr)
		found := false
		for scanner.Scan() {
			if scanner.Text() == "waiting" {
				found = true
				observed <- true
				break
			}
		}
		if !found {
			observed <- false
		}
	}()
	select {
	case found := <-observed:
		if !found {
			cancel()
			cmd.Wait()
			t.Fatal("command did not enter held work")
		}
	case <-guard.Done():
		cmd.Wait()
		t.Fatal("command readiness deadline")
	}
	if err := cmd.Process.Signal(os.Interrupt); err != nil {
		cancel()
		cmd.Wait()
		t.Fatal(err)
	}
	err = cmd.Wait()
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 2 || stdout.String() != "accepted 1\n" {
		t.Fatalf("exit=%v stdout=%q", err, stdout.String())
	}
	data, err := os.ReadFile(receipt)
	if err != nil || string(data) != "accepted=1\n" {
		t.Fatalf("receipt=%q err=%v", data, err)
	}
}

func TestRequiredReceiptFailureProcess(t *testing.T) {
	binary := buildCommand(t)
	receipt := filepath.Join(t.TempDir(), "absent", "receipt")
	guard, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(guard, binary, "--receipt", receipt, "1")
	cmd.Env = []string{"PATH=" + os.Getenv("PATH")}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 2 || stdout.String() != "accepted 1\n" || strings.TrimSpace(stderr.String()) == "" {
		t.Fatalf("err=%v stdout=%q stderr=%q", err, stdout.String(), stderr.String())
	}
	if _, err := os.Stat(receipt); !os.IsNotExist(err) {
		t.Fatalf("unexpected receipt status: %v", err)
	}
}
