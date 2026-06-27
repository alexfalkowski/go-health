package checker

import (
	"net/http"
	"slices"

	"github.com/alexfalkowski/go-health/v2/net"
)

// Option configures checkers created by the constructors in this package.
//
// The supported options depend on the checker type. Unsupported options are
// ignored by constructors that do not use them.
type Option interface {
	apply(opts *options)
}

type options struct {
	roundTripper http.RoundTripper
	dialer       net.Dialer
	headers      http.Header
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

// WithHeader adds a header value used by HTTPChecker.
func WithHeader(name, value string) Option {
	return optionFunc(func(o *options) {
		if o.headers == nil {
			o.headers = make(http.Header)
		}

		o.headers.Add(name, value)
	})
}

// WithDialer sets the dialer used by TCPChecker.
func WithDialer(dialer net.Dialer) Option {
	return optionFunc(func(o *options) {
		o.dialer = dialer
	})
}

// WithURLs sets the list of URLs used by OnlineChecker.
//
// Providing at least one URL replaces the package defaults entirely. Providing
// no URLs leaves the default connectivity endpoint list in place.
func WithURLs(urls ...string) Option {
	return optionFunc(func(o *options) {
		o.urls = slices.Clone(urls)
	})
}

func parseOptions(opts ...Option) *options {
	parsed := &options{}
	for _, o := range opts {
		o.apply(parsed)
	}
	if parsed.roundTripper == nil {
		parsed.roundTripper = http.DefaultTransport
	}
	if parsed.dialer == nil {
		parsed.dialer = net.DefaultDialer
	}
	if len(parsed.urls) == 0 {
		parsed.urls = []string{
			"https://google.com/generate_204",
			"https://cp.cloudflare.com/generate_204",
			"https://connectivity-check.ubuntu.com",
		}
	}

	return parsed
}
