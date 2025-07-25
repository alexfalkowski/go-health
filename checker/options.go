package checker

import (
	"net/http"

	"github.com/alexfalkowski/go-health/v2/net"
)

// Option for checker.
type Option interface {
	apply(opts *options)
}

type options struct {
	roundTripper http.RoundTripper
	dialer       net.Dialer
}

type optionFunc func(*options)

func (f optionFunc) apply(o *options) {
	f(o)
}

// WithRoundTripper for checker.
func WithRoundTripper(rt http.RoundTripper) Option {
	return optionFunc(func(o *options) {
		o.roundTripper = rt
	})
}

// WithDialer for checker.
func WithDialer(dialer net.Dialer) Option {
	return optionFunc(func(o *options) {
		o.dialer = dialer
	})
}

func parseOptions(opts ...Option) *options {
	os := &options{}
	for _, o := range opts {
		o.apply(os)
	}
	if os.roundTripper == nil {
		os.roundTripper = http.DefaultTransport
	}
	if os.dialer == nil {
		os.dialer = net.DefaultDialer
	}

	return os
}
