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
	urls         []string
}

type optionFunc func(*options)

func (f optionFunc) apply(o *options) {
	f(o)
}

// WithRoundTripper sets the HTTP transport used by HTTPChecker and OnlineChecker.
func WithRoundTripper(rt http.RoundTripper) Option {
	return optionFunc(func(o *options) {
		o.roundTripper = rt
	})
}

// WithDialer sets the dialer used by TCPChecker.
func WithDialer(dialer net.Dialer) Option {
	return optionFunc(func(o *options) {
		o.dialer = dialer
	})
}

// WithURLs sets the list of URLs used by OnlineChecker.
func WithURLs(urls ...string) Option {
	return optionFunc(func(o *options) {
		o.urls = urls
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
	if len(os.urls) == 0 {
		os.urls = []string{
			"https://google.com/generate_204",
			"https://cp.cloudflare.com/generate_204",
			"https://connectivity-check.ubuntu.com",
		}
	}

	return os
}
