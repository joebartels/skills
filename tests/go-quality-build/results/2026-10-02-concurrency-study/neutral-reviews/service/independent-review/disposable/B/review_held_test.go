package workers

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type probeGroup struct {
	mu         sync.Mutex
	leases     []*probeLease
	earlyClose atomic.Bool
}
type probeLease struct {
	group                              *probeGroup
	started, stopped, finished, closed chan struct{}
	release                            <-chan struct{}
	fault                              <-chan struct{}
	runError, closeError               error
	expectRun                          bool
	closes                             atomic.Int64
}

func (g *probeGroup) add(l *probeLease) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.leases = append(g.leases, l)
}
func (l *probeLease) Run(ctx context.Context, j Job) error {
	close(l.started)
	defer close(l.finished)
	if l.runError != nil {
		<-l.fault
		return l.runError
	}
	<-ctx.Done()
	close(l.stopped)
	<-l.release
	return ctx.Err()
}
func (l *probeLease) Close() error {
	l.group.mu.Lock()
	for _, other := range l.group.leases {
		if other.expectRun {
			select {
			case <-other.finished:
			default:
				l.group.earlyClose.Store(true)
			}
		}
	}
	l.group.mu.Unlock()
	if l.closes.Add(1) == 1 {
		close(l.closed)
	}
	return l.closeError
}
func probeWait(t *testing.T, ch <-chan struct{}, what string) bool {
	t.Helper()
	select {
	case <-ch:
		return true
	case <-time.After(time.Second):
		t.Errorf("missing event: %s", what)
		return false
	}
}
func TestControllerPartialStartAndBlockedAdmission(t *testing.T) {
	for _, mode := range []string{"startup failure", "worker failure", "caller cancellation", "acquired before stop"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			release := make(chan struct{})
			fault := make(chan struct{})
			abort := make(chan struct{})
			attemptThird := make(chan struct{})
			laterOpen := make(chan struct{}, 1)
			finished := make(chan struct{})
			result := make(chan error, 1)
			jobs := make(chan Job)
			feedDone := make(chan struct{})
			var group probeGroup
			var releaseOnce, faultOnce, abortOnce sync.Once
			closeFailure := errors.New("independent close failure")
			startupFailure := errors.New("independent startup failure")
			workerFailure := errors.New("independent worker failure")
			makeLease := func(expect bool) *probeLease {
				return &probeLease{group: &group, started: make(chan struct{}), stopped: make(chan struct{}), finished: make(chan struct{}), closed: make(chan struct{}), release: release, fault: fault, expectRun: expect}
			}
			first := makeLease(true)
			first.closeError = closeFailure
			second := makeLease(mode != "acquired before stop")
			if mode == "worker failure" {
				second.runError = workerFailure
			}
			t.Cleanup(func() {
				cancel()
				releaseOnce.Do(func() { close(release) })
				faultOnce.Do(func() { close(fault) })
				abortOnce.Do(func() { close(abort) })
				if !probeWait(t, feedDone, "feeder cleanup") {
					return
				}
				probeWait(t, finished, "host cleanup")
				for _, lease := range []*probeLease{first, second} {
					select {
					case <-lease.started:
						probeWait(t, lease.finished, "owned test callback cleanup")
					default:
					}
				}
			})
			go func() {
				defer close(feedDone)
				defer close(jobs)
				for _, job := range []Job{1, 2} {
					select {
					case jobs <- job:
					case <-abort:
						return
					}
				}
				close(attemptThird)
				select {
				case jobs <- 3:
				case <-abort:
					return
				}
				<-abort
			}()
			go func() {
				defer close(finished)
				result <- Serve(ctx, jobs, 2, func(ctx context.Context, j Job) (Lease, error) {
					switch j {
					case 1:
						group.add(first)
						return first, nil
					case 2:
						<-first.started
						if mode == "startup failure" {
							<-fault
							return nil, startupFailure
						}
						group.add(second)
						if mode == "acquired before stop" {
							cancel()
						}
						return second, nil
					default:
						select {
						case laterOpen <- struct{}{}:
						default:
						}
						return nil, errors.New("unexpected later open")
					}
				})
			}()
			if !probeWait(t, first.started, "first run started") || !probeWait(t, attemptThird, "blocked third admission attempted") {
				return
			}
			if mode == "caller cancellation" {
				if !probeWait(t, second.started, "second run started") {
					return
				}
				cancel()
			} else {
				faultOnce.Do(func() { close(fault) })
			}
			if !probeWait(t, first.stopped, "first run requested to stop") {
				return
			}
			if group.earlyClose.Load() || first.closes.Load() != 0 || second.closes.Load() != 0 {
				t.Error("lease closed before all started users joined")
			}
			select {
			case <-finished:
				t.Error("host returned before held cleanup completed")
			default:
			}
			select {
			case <-laterOpen:
				t.Error("later job opened despite bounded/stopped admission")
			default:
			}
			if mode == "acquired before stop" {
				select {
				case <-second.started:
					t.Error("Run started on lease acquired after stopping")
				default:
				}
			}
			releaseOnce.Do(func() { close(release) })
			if !probeWait(t, finished, "host completion after release") {
				return
			}
			err := <-result
			if !errors.Is(err, closeFailure) {
				t.Errorf("close failure lost: %v", err)
			}
			if mode == "startup failure" && !errors.Is(err, startupFailure) {
				t.Errorf("startup failure lost: %v", err)
			}
			if mode == "worker failure" && !errors.Is(err, workerFailure) {
				t.Errorf("worker failure lost: %v", err)
			}
			if (mode == "caller cancellation" || mode == "acquired before stop") && !errors.Is(err, context.Canceled) {
				t.Errorf("cancellation lost: %v", err)
			}
			if first.closes.Load() != 1 {
				t.Errorf("first closes=%d", first.closes.Load())
			}
			if mode != "startup failure" && second.closes.Load() != 1 {
				t.Errorf("second closes=%d", second.closes.Load())
			}
			if group.earlyClose.Load() {
				t.Error("lease closed before all started users joined")
			}
			select {
			case <-laterOpen:
				t.Error("later admission after failure/cancellation")
			default:
			}
		})
	}
}

type capacityLease struct {
	alive, active, accepted, closed *atomic.Int64
	early                           *atomic.Bool
	started, closeEntered           chan<- struct{}
	release, closeRelease           <-chan struct{}
	finished                        atomic.Bool
}

func (l *capacityLease) Run(context.Context, Job) error {
	l.active.Add(1)
	l.started <- struct{}{}
	<-l.release
	l.accepted.Add(1)
	l.active.Add(-1)
	l.finished.Store(true)
	return nil
}
func (l *capacityLease) Close() error {
	if !l.finished.Load() || l.active.Load() != 0 {
		l.early.Store(true)
	}
	l.closeEntered <- struct{}{}
	<-l.closeRelease
	l.alive.Add(-1)
	l.closed.Add(1)
	return nil
}
func TestControllerCapacityAndAcceptedWork(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	jobs := make(chan Job, 5)
	for i := 1; i <= 5; i++ {
		jobs <- Job(i)
	}
	close(jobs)
	release, closeRelease := make(chan struct{}), make(chan struct{})
	var releaseOnce, closeOnce sync.Once
	started, closeEntered := make(chan struct{}, 5), make(chan struct{}, 5)
	thirdOpen := make(chan struct{}, 1)
	var alive, maximum, active, accepted, closed, opened atomic.Int64
	var early atomic.Bool
	done := make(chan struct{})
	result := make(chan error, 1)
	t.Cleanup(func() {
		cancel()
		releaseOnce.Do(func() { close(release) })
		closeOnce.Do(func() { close(closeRelease) })
		probeWait(t, done, "capacity host cleanup")
	})
	go func() {
		defer close(done)
		result <- Serve(ctx, jobs, 2, func(context.Context, Job) (Lease, error) {
			if opened.Add(1) == 3 {
				thirdOpen <- struct{}{}
			}
			n := alive.Add(1)
			for old := maximum.Load(); n > old; old = maximum.Load() {
				if maximum.CompareAndSwap(old, n) {
					break
				}
			}
			return &capacityLease{alive: &alive, active: &active, accepted: &accepted, closed: &closed, early: &early, started: started, closeEntered: closeEntered, release: release, closeRelease: closeRelease}, nil
		})
	}()
	for i := 0; i < 2; i++ {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Error("available capacity not used concurrently")
			return
		}
	}
	if maximum.Load() > 2 {
		t.Errorf("acquired leases exceed2: %d", maximum.Load())
	}
	releaseOnce.Do(func() { close(release) })
	select {
	case <-closeEntered:
	case <-time.After(time.Second):
		t.Error("close not entered after runs completed")
		return
	}
	select {
	case <-thirdOpen:
		t.Error("next lease acquired while Close still holds capacity")
	case <-time.After(50 * time.Millisecond):
	}
	if alive.Load() != 2 || maximum.Load() > 2 {
		t.Errorf("held Close capacity alive=%d max=%d", alive.Load(), maximum.Load())
	}
	closeOnce.Do(func() { close(closeRelease) })
	if !probeWait(t, done, "finite jobs completed") {
		return
	}
	if err := <-result; err != nil {
		t.Errorf("unexpected failure: %v", err)
	}
	if maximum.Load() > 2 || alive.Load() != 0 || accepted.Load() != 5 || closed.Load() != 5 || early.Load() {
		t.Errorf("max=%d alive=%d accepted=%d closed=%d early=%v", maximum.Load(), alive.Load(), accepted.Load(), closed.Load(), early.Load())
	}
}

type inputProbeLease struct{ started chan struct{} }

func (l *inputProbeLease) Run(context.Context, Job) error { close(l.started); return nil }
func (l *inputProbeLease) Close() error                   { return nil }
func TestControllerCancellationWithOpenInput(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	jobs := make(chan Job, 1)
	jobs <- 1
	started := make(chan struct{})
	done := make(chan struct{})
	result := make(chan error, 1)
	var opens atomic.Int64
	t.Cleanup(func() { cancel(); close(jobs); probeWait(t, done, "open-input host cleanup") })
	go func() {
		defer close(done)
		result <- Serve(ctx, jobs, 2, func(context.Context, Job) (Lease, error) { opens.Add(1); return &inputProbeLease{started}, nil })
	}()
	if !probeWait(t, started, "first work entered before in-flight cancellation") {
		return
	}
	cancel()
	if !probeWait(t, done, "blocked admission observes cancellation before input closure") {
		return
	}
	if err := <-result; !errors.Is(err, context.Canceled) {
		t.Errorf("cancellation lost: %v", err)
	}
	if opens.Load() != 1 {
		t.Errorf("opened=%d", opens.Load())
	}
}
