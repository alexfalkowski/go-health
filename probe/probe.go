package probe

import (
	"context"
	"sync"
	"time"

	"github.com/alexfalkowski/go-health/v2/checker"
)

// NewProbe returns a Probe that runs ch at the given period.
func NewProbe(name string, period time.Duration, ch checker.Checker) *Probe {
	return &Probe{name: name, period: period, checker: ch, ticker: nil, ch: nil, done: nil, mux: sync.Mutex{}}
}

// Probe periodically runs a Checker and emits ticks.
type Probe struct {
	checker checker.Checker
	ticker  *time.Ticker
	ch      chan *Tick
	done    chan struct{}
	name    string
	period  time.Duration
	mux     sync.Mutex
	wg      sync.WaitGroup
}

// Start begins running checks and returns a channel of ticks.
//
// Start performs an initial check before starting the periodic checks.
func (p *Probe) Start() <-chan *Tick {
	p.mux.Lock()
	defer p.mux.Unlock()

	done := make(chan struct{})
	ch := make(chan *Tick, 1)
	ticker := time.NewTicker(p.period)

	p.done = done
	p.ch = ch
	p.ticker = ticker

	// Check on startup.
	p.tick(ch, done)
	p.wg.Go(func() {
		p.start(ch, done, ticker)
	})

	return ch
}

// Stop stops the probe.
func (p *Probe) Stop() {
	p.mux.Lock()
	defer p.mux.Unlock()

	if p.done == nil {
		return
	}

	done := p.done
	ticker := p.ticker

	p.done = nil
	p.ch = nil
	p.ticker = nil

	ticker.Stop()
	close(done)
	p.wg.Wait()
}

func (p *Probe) start(ch chan *Tick, done <-chan struct{}, ticker *time.Ticker) {
	defer close(ch)

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			p.tick(ch, done)
		}
	}
}

func (p *Probe) tick(ch chan<- *Tick, done <-chan struct{}) {
	tick := NewTick(p.name, p.checker.Check(context.Background()))

	select {
	case <-done:
	case ch <- tick:
	}
}
