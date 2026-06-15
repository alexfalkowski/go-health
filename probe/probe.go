package probe

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alexfalkowski/go-health/v2/checker"
	"github.com/alexfalkowski/go-sync"
)

// ErrInvalidPeriod is reported when a probe is configured with a period less
// than or equal to zero.
var ErrInvalidPeriod = errors.New("health: invalid period")

// NewProbe returns a Probe that runs check at the given period.
//
// The check must be non-nil. The probe does not start until Start is called.
func NewProbe(name string, period time.Duration, check checker.Checker) *Probe {
	return &Probe{name: name, period: period, checker: check, mux: sync.Mutex{}}
}

// Probe periodically runs a Checker and emits ticks.
type Probe struct {
	checker checker.Checker
	active  *activeProbe
	name    string
	period  time.Duration
	mux     sync.Mutex
}

type activeProbe struct {
	ticker *time.Ticker
	ch     chan *Tick
	ready  chan struct{}
	cancel context.CancelFunc
	wg     *sync.WaitGroup
}

// Start begins running checks and returns a channel of ticks.
//
// Start performs an initial check with ctx before returning so callers can
// observe an immediate result. The context controls startup and the initial
// readiness wait; if it is canceled before startup completes, Start returns the
// context error and no tick channel. If Stop cancels startup before the initial
// result can be emitted, Start can return nil error with a closed tick channel
// and no initial tick. Canceling ctx after Start returns does not stop the
// probe; use Stop to end the probe lifecycle. Keep receiving from the returned
// channel while the probe is running. If the period is zero or negative, Start
// returns a closed channel containing one tick whose error wraps
// ErrInvalidPeriod. If the probe is already running and ctx remains valid
// through readiness, Start returns the existing channel.
func (p *Probe) Start(ctx context.Context) (<-chan *Tick, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	ch, ready, started := p.ensureStarted(ctx)
	if ready != nil {
		select {
		case <-ctx.Done():
			if started {
				_ = p.Stop(context.WithoutCancel(ctx))
			}
			return nil, ctx.Err()
		case <-ready:
			if err := ctx.Err(); err != nil {
				if started {
					_ = p.Stop(context.WithoutCancel(ctx))
				}
				return nil, err
			}
		}
	}
	return ch, nil
}

func (p *Probe) ensureStarted(ctx context.Context) (<-chan *Tick, <-chan struct{}, bool) {
	p.mux.Lock()
	defer p.mux.Unlock()

	if p.active != nil {
		return p.active.ch, p.active.ready, false
	}

	ch := make(chan *Tick, 1)
	ready := make(chan struct{})

	if p.period <= 0 {
		ch <- NewTick(p.name, fmt.Errorf("%w: %s", ErrInvalidPeriod, p.period))
		close(ch)
		close(ready)
		return ch, ready, false
	}

	runCtx, cancel := context.WithCancel(context.Background())
	ticker := time.NewTicker(p.period)
	wg := &sync.WaitGroup{}

	p.active = &activeProbe{ch: ch, ready: ready, ticker: ticker, cancel: cancel, wg: wg}
	wg.Go(func() {
		p.start(runCtx, ctx, ch, ticker, ready)
	})

	return ch, ready, true
}

// Stop stops the probe.
//
// Stop is safe to call before Start and safe to call multiple times. It cancels
// the context for any in-flight check, closes the tick channel once the worker
// exits, and waits for the probe goroutine to finish. If ctx expires while
// waiting, Stop returns ctx.Err().
func (p *Probe) Stop(ctx context.Context) error {
	p.mux.Lock()

	active := p.active
	if active == nil {
		p.mux.Unlock()
		return nil
	}

	active.cancel()
	active.ticker.Stop()
	p.mux.Unlock()

	done := make(chan struct{})
	go func() {
		active.wg.Wait()
		p.clearActive(active)
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

func (p *Probe) clearActive(active *activeProbe) {
	p.mux.Lock()
	defer p.mux.Unlock()

	if p.active == active {
		p.active = nil
	}
}

func (p *Probe) start(runCtx, startCtx context.Context, ch chan *Tick, ticker *time.Ticker, ready chan<- struct{}) {
	defer close(ch)

	// Check on startup before periodic checks begin.
	initialCtx, cancel := context.WithCancel(runCtx)
	stop := context.AfterFunc(startCtx, cancel)
	if !p.tick(initialCtx, ch) {
		stop()
		cancel()
		close(ready)
		return
	}
	stop()
	cancel()
	close(ready)

	for {
		select {
		case <-runCtx.Done():
			return
		case <-ticker.C:
			if !p.tick(runCtx, ch) {
				return
			}
		}
	}
}

func (p *Probe) tick(ctx context.Context, ch chan<- *Tick) bool {
	// Run the check first so we can observe cancellation that happens while Check is blocked.
	err := p.checker.Check(ctx)
	// If Stop canceled the context during Check, drop the stale result instead of emitting a tick.
	if ctx.Err() != nil {
		return false
	}

	select {
	case <-ctx.Done():
		return false
	case ch <- NewTick(p.name, err):
		return true
	}
}
