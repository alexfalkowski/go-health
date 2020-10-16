package prb

import (
	"context"
	"time"

	"github.com/alexfalkowski/go-health/pkg/chk"
)

// NewProbe with period and checker.
func NewProbe(name string, period time.Duration, checker chk.Checker) *Probe {
	return &Probe{name: name, period: period, checker: checker}
}

// Probe is a periodic checker.
type Probe struct {
	name    string
	period  time.Duration
	ticker  *time.Ticker
	checker chk.Checker
	ch      chan *Tick
	done    chan struct{}
}

// Start the probe.
func (p *Probe) Start() <-chan *Tick {
	p.done = make(chan struct{}, 1)
	p.ch = make(chan *Tick, 1)
	p.ticker = time.NewTicker(p.period)

	go p.start()

	return p.ch
}

// Stop the probe.
func (p *Probe) Stop() {
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
			p.ch <- NewTick(p.name, p.checker.Check(context.Background()))
		}
	}
}
