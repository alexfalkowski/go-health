// Package probe provides periodic execution of checkers.
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
}

// Start begins running checks and returns a channel of ticks.
//
// Start performs an initial check before starting the periodic checks.
func (p *Probe) Start() <-chan *Tick {
	p.mux.Lock()
	defer p.mux.Unlock()

	p.done = make(chan struct{}, 1)
	p.ch = make(chan *Tick, 1)
	p.ticker = time.NewTicker(p.period)

	// Check on startup.
	p.tick()
	go p.start()

	return p.ch
}

// Stop stops the probe.
func (p *Probe) Stop() {
	p.mux.Lock()
	defer p.mux.Unlock()

	p.ticker.Stop()
	close(p.done)
}

func (p *Probe) start() {
	defer close(p.ch)

	for {
		select {
		case <-p.done:
			return
		case <-p.ticker.C:
			p.tick()
		}
	}
}

func (p *Probe) tick() {
	p.ch <- NewTick(p.name, p.checker.Check(context.Background()))
}
