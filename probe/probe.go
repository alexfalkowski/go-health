package probe

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alexfalkowski/go-health/v2/checker"
	"github.com/alexfalkowski/go-sync"
)

// ErrInvalidPeriod when the probe period is not greater than zero.
var ErrInvalidPeriod = errors.New("health: invalid period")

// NewProbe returns a Probe that runs ch at the given period.
func NewProbe(name string, period time.Duration, ch checker.Checker) *Probe {
	return &Probe{name: name, period: period, checker: ch, ticker: nil, ch: nil, mux: sync.Mutex{}}
}

// Probe periodically runs a Checker and emits ticks.
type Probe struct {
	checker checker.Checker
	ticker  *time.Ticker
	ch      chan *Tick
	cancel  context.CancelFunc
	name    string
	period  time.Duration
	mux     sync.Mutex
	wg      sync.WaitGroup
}

// Start begins running checks and returns a channel of ticks.
//
// Start performs an initial check before starting the periodic checks.
func (p *Probe) Start() <-chan *Tick {
	ch, ready := p.ensureStarted()
	if ready != nil {
		<-ready
	}
	return ch
}

func (p *Probe) ensureStarted() (<-chan *Tick, <-chan struct{}) {
	p.mux.Lock()
	defer p.mux.Unlock()

	if p.cancel != nil {
		return p.ch, nil
	}

	ch := make(chan *Tick, 1)
	ready := make(chan struct{})

	if p.period <= 0 {
		ch <- NewTick(p.name, fmt.Errorf("%w: %s", ErrInvalidPeriod, p.period))
		close(ch)
		close(ready)
		return ch, ready
	}

	ctx, cancel := context.WithCancel(context.Background())
	ticker := time.NewTicker(p.period)

	p.ch = ch
	p.ticker = ticker
	p.cancel = cancel

	p.wg.Go(func() {
		p.start(ctx, ch, ticker, ready)
	})

	return ch, ready
}

// Stop stops the probe.
func (p *Probe) Stop() {
	p.mux.Lock()
	defer p.mux.Unlock()

	if p.cancel == nil {
		return
	}

	ticker := p.ticker
	cancel := p.cancel

	p.ch = nil
	p.ticker = nil
	p.cancel = nil

	cancel()
	ticker.Stop()
	p.wg.Wait()
}

func (p *Probe) start(ctx context.Context, ch chan *Tick, ticker *time.Ticker, ready chan<- struct{}) {
	defer close(ch)

	// Check on startup before periodic checks begin.
	if !p.tick(ctx, ch) {
		close(ready)
		return
	}
	close(ready)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !p.tick(ctx, ch) {
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
