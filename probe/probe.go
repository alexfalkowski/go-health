package probe

import (
	"context"
	"sync"
	"time"

	"github.com/alexfalkowski/go-health/checker"
)

// NewProbe with period and checker.
func NewProbe(name string, period time.Duration, ch checker.Checker) *Probe {
	return &Probe{name: name, period: period, checker: ch, ticker: nil, ch: nil, done: nil, mux: sync.Mutex{}}
}

// Probe is a periodic checker.
type Probe struct {
	checker checker.Checker
	ticker  *time.Ticker
	ch      chan *Tick
	done    chan struct{}
	name    string
	period  time.Duration
	mux     sync.Mutex
}

// Start the probe.
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

// Stop the probe.
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
