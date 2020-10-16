package svr

import (
	"context"
	"time"

	"github.com/alexfalkowski/go-health/pkg/chk"
)

// Probe is a periodic checker.
type Probe struct {
	name    string
	period  time.Duration
	ticker  *time.Ticker
	checker chk.Checker
	ch      chan *ProbeTick
}

// NewProbe with period and checker.
func NewProbe(name string, period time.Duration, checker chk.Checker) *Probe {
	return &Probe{name: name, period: period, checker: checker}
}

// Start the probe.
func (p *Probe) Start(ctx context.Context) chan *ProbeTick {
	p.ch = make(chan *ProbeTick)
	p.ticker = time.NewTicker(p.period)

	go p.start(ctx)

	return p.ch
}

func (p *Probe) start(ctx context.Context) {
	defer close(p.ch)
	defer p.ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-p.ticker.C:
			p.ch <- NewProbeTick(p.name, p.checker.Check(ctx))
		}
	}
}
