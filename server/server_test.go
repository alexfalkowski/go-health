package server_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"testing/synctest"
	"time"

	"github.com/alexfalkowski/go-health/v2/checker"
	"github.com/alexfalkowski/go-health/v2/internal/test"
	testchecker "github.com/alexfalkowski/go-health/v2/internal/test/checker"
	"github.com/alexfalkowski/go-health/v2/internal/test/sql"
	testsubscriber "github.com/alexfalkowski/go-health/v2/internal/test/subscriber"
	"github.com/alexfalkowski/go-health/v2/net"
	"github.com/alexfalkowski/go-health/v2/probe"
	"github.com/alexfalkowski/go-health/v2/server"
	"github.com/stretchr/testify/require"
)

const (
	timeout = 2 * time.Second
	period  = 500 * time.Millisecond
	wait    = 1 * time.Second
)

var invalidURL = string([]byte{0x7f})

func TestDoubleStart(t *testing.T) {
	s := server.NewServer()
	defer func() { _ = s.Stop(context.Background()) }()

	checker := checker.NewHTTPChecker(
		"https://www.google.com/",
		timeout,
		checker.WithRoundTripper(http.DefaultTransport),
	)
	r := server.NewRegistration("google", period, checker)
	s.Register("test", r)

	_ = s.Observe("test", "livez", r.Name)
	_, _ = s.Observer("test", "livez")

	_ = s.Observe("test", "livez", r.Name)
	ob, _ := s.Observer("test", "livez")

	_ = s.Start(context.Background())
	_ = s.Start(context.Background())

	testsubscriber.RequireObserverNoError(t, ob)
}

func TestStartWithCanceledContextReturnsContextError(t *testing.T) {
	s := server.NewServer()
	registration := server.NewRegistration("noop", time.Hour, checker.NewNoopChecker())
	s.Register("test", registration)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	err := s.Start(ctx)

	require.ErrorIs(t, err, context.Canceled)
	require.NoError(t, s.Start(t.Context()))
	t.Cleanup(func() { _ = s.Stop(context.Background()) })
}

func TestStartReturnsServiceStartError(t *testing.T) {
	s := server.NewServer()
	ch := testchecker.NewBlockingChecker()

	registration := server.NewRegistration("blocking", time.Hour, ch)
	s.Register("test", registration)

	ctx, cancel := context.WithCancel(t.Context())
	errc := make(chan error, 1)
	go func() {
		errc <- s.Start(ctx)
	}()

	testchecker.WaitForStarted(t, ch.Started)
	cancel()

	require.ErrorIs(t, <-errc, context.Canceled)
	testchecker.WaitForCanceled(t, ch.Canceled)

	registration = server.NewRegistration("blocking", time.Hour, checker.NewNoopChecker())
	s.Register("test", registration)
	require.NoError(t, s.Start(t.Context()))
	t.Cleanup(func() { _ = s.Stop(context.Background()) })
}

func TestStopReturnsContextError(t *testing.T) {
	s := server.NewServer()
	ch := testchecker.NewBlockingPeriodicChecker()

	registration := server.NewRegistration("blocking", time.Millisecond, ch)
	s.Register("test", registration)
	require.NoError(t, s.Start(t.Context()))
	testchecker.WaitForStarted(t, ch.PeriodicStarted)
	t.Cleanup(func() {
		close(ch.Release)
		_ = s.Stop(context.Background())
	})

	ctx, cancel := context.WithTimeout(t.Context(), time.Millisecond)
	defer cancel()

	require.ErrorIs(t, s.Stop(ctx), context.DeadlineExceeded)
}

func TestOnlineChecker(t *testing.T) {
	s := server.NewServer()
	defer func() { _ = s.Stop(context.Background()) }()

	r := server.NewOnlineRegistration(0, period)
	s.Register("test", r)

	_ = s.Observe("test", "livez", r.Name)
	ob, _ := s.Observer("test", "livez")

	_ = s.Start(t.Context())

	testsubscriber.RequireObserverNoError(t, ob)
}

func TestInvalidOnlineChecker(t *testing.T) {
	s := server.NewServer()
	defer func() { _ = s.Stop(context.Background()) }()

	r := server.NewOnlineRegistration(0, period, checker.WithURLs(invalidURL, "https://www.assaaasss.com/"))
	s.Register("test", r)

	_ = s.Observe("test", "livez", r.Name)
	ob, _ := s.Observer("test", "livez")

	_ = s.Start(t.Context())

	testsubscriber.RequireObserverError(t, ob)
}

func TestValidHTTPChecker(t *testing.T) {
	s := server.NewServer()
	defer func() { _ = s.Stop(context.Background()) }()

	checker := checker.NewHTTPChecker("https://www.google.com/", 0)
	r := server.NewRegistration("google", period, checker)
	s.Register("test", r)

	_ = s.Observe("test", "livez", r.Name)
	ob, _ := s.Observer("test", "livez")

	_ = s.Start(t.Context())

	testsubscriber.RequireObserverNoError(t, ob)
}

func TestInvalidURLHTTPChecker(t *testing.T) {
	s := server.NewServer()
	defer func() { _ = s.Stop(context.Background()) }()

	checker := checker.NewHTTPChecker("https://www.assaaasss.com/", timeout)
	r := server.NewRegistration("assaaasss", period, checker)
	s.Register("test", r)

	_ = s.Observe("test", "livez", r.Name)
	ob, _ := s.Observer("test", "livez")

	_ = s.Start(t.Context())

	testsubscriber.RequireObserverError(t, ob)
}

func TestMalformedURLHTTPChecker(t *testing.T) {
	s := server.NewServer()
	defer func() { _ = s.Stop(context.Background()) }()

	checker := checker.NewHTTPChecker(invalidURL, timeout)
	r := server.NewRegistration("assaaasss", period, checker)
	s.Register("test", r)

	_ = s.Observe("test", "livez", r.Name)
	ob, _ := s.Observer("test", "livez")

	_ = s.Start(t.Context())

	testsubscriber.RequireObserverError(t, ob)
}

func TestInvalidCodeHTTPChecker(t *testing.T) {
	s := server.NewServer()
	defer func() { _ = s.Stop(context.Background()) }()

	checker := checker.NewHTTPChecker(test.StatusURL("400"), timeout)
	r := server.NewRegistration("http400", period, checker)
	s.Register("test", r)

	_ = s.Observe("test", "livez", r.Name)
	ob, _ := s.Observer("test", "livez")

	_ = s.Start(t.Context())

	testsubscriber.RequireObserverError(t, ob)
}

func TestTimeoutHTTPChecker(t *testing.T) {
	s := server.NewServer()
	defer func() { _ = s.Stop(context.Background()) }()

	checker := checker.NewHTTPChecker(test.StatusURL("200?sleep=5s"), timeout)
	r := server.NewRegistration("http200", period, checker)
	s.Register("test", r)

	_ = s.Observe("test", "livez", r.Name)
	ob, _ := s.Observer("test", "livez")

	_ = s.Start(t.Context())

	testsubscriber.RequireObserverError(t, ob)
}

func TestInvalidPeriod(t *testing.T) {
	s := server.NewServer()
	defer func() { _ = s.Stop(context.Background()) }()

	registration := server.NewRegistration("noop", 0, checker.NewNoopChecker())
	s.Register("test", registration)

	require.NoError(t, s.Observe("test", "livez", registration.Name))
	ob, _ := s.Observer("test", "livez")

	require.NotPanics(t, func() {
		_ = s.Start(t.Context())
	})

	require.Eventually(t, func() bool {
		return ob.Error() != nil
	}, time.Second, 10*time.Millisecond)
	require.ErrorIs(t, ob.Error(), probe.ErrInvalidPeriod)
}

func TestValidTCPChecker(t *testing.T) {
	s := server.NewServer()
	defer func() { _ = s.Stop(context.Background()) }()

	checker := checker.NewTCPChecker(
		"www.google.com:80",
		timeout,
		checker.WithDialer(net.DefaultDialer),
	)
	r := server.NewRegistration("tcp-google", period, checker)
	s.Register("test", r)

	_ = s.Observe("test", "livez", r.Name)
	ob, _ := s.Observer("test", "livez")

	_ = s.Start(t.Context())

	testsubscriber.RequireObserverNoError(t, ob)
}

func TestInvalidAddressTCPChecker(t *testing.T) {
	s := server.NewServer()
	defer func() { _ = s.Stop(context.Background()) }()

	checker := checker.NewTCPChecker("www.assaaasss.com:80", timeout)
	r := server.NewRegistration("tcp-assaaasss", period, checker)
	s.Register("test", r)

	_ = s.Observe("test", "livez", r.Name)
	ob, _ := s.Observer("test", "livez")

	_ = s.Start(t.Context())

	testsubscriber.RequireObserverError(t, ob)
	require.Error(t, ob.Errors()["tcp-assaaasss"])
}

func TestValidDBChecker(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		s := server.NewServer()
		t.Cleanup(func() { _ = s.Stop(context.Background()) })

		checker := checker.NewDBChecker(sql.OKPinger{}, timeout)
		r := server.NewRegistration("db", period, checker)
		s.Register("test", r)

		_ = s.Observe("test", "livez", r.Name)
		ob, _ := s.Observer("test", "livez")

		_ = s.Start(t.Context())
		synctest.Wait()

		require.NoError(t, ob.Error())
	})
}

//nolint:err113
func TestInvalidDBChecker(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		s := server.NewServer()
		t.Cleanup(func() { _ = s.Stop(context.Background()) })

		errPing := errors.New("ping failed")
		checker := checker.NewDBChecker(sql.ErrorPinger{Err: errPing}, timeout)
		r := server.NewRegistration("db", period, checker)
		s.Register("test", r)

		_ = s.Observe("test", "livez", r.Name)
		ob, _ := s.Observer("test", "livez")

		_ = s.Start(t.Context())
		synctest.Wait()

		require.Error(t, ob.Error())
		require.ErrorIs(t, ob.Error(), errPing)
		require.ErrorIs(t, ob.Errors()["db"], errPing)
	})
}

//nolint:err113
func TestValidReadyChecker(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		s := server.NewServer()
		t.Cleanup(func() { _ = s.Stop(context.Background()) })

		errNotReady := errors.New("not ready")
		checker := checker.NewReadyChecker(errNotReady)
		r := server.NewRegistration("ready", period, checker)
		s.Register("test", r)

		_ = s.Observe("test", "livez", r.Name)
		ob, _ := s.Observer("test", "livez")

		_ = s.Start(t.Context())
		synctest.Wait()

		require.Error(t, ob.Error())

		checker.Ready()
		time.Sleep(period)
		synctest.Wait()

		require.NoError(t, ob.Error())
	})
}

func TestValidNoopChecker(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		s := server.NewServer()
		t.Cleanup(func() { _ = s.Stop(context.Background()) })

		checker := checker.NewNoopChecker()
		r := server.NewRegistration("noop", period, checker)
		s.Register("test", r)

		_ = s.Observe("test", "livez", r.Name)
		ob, _ := s.Observer("test", "livez")

		_ = s.Start(t.Context())
		synctest.Wait()

		require.NoError(t, ob.Error())
	})
}

func TestInvalidObserver(t *testing.T) {
	s := server.NewServer()
	defer func() { _ = s.Stop(context.Background()) }()

	cc := checker.NewHTTPChecker(test.StatusURL("400"), timeout)
	hr := server.NewRegistration("http1", period, cc)
	tc := checker.NewTCPChecker("google.com:9000", timeout)
	tr := server.NewRegistration("tcp1", period, tc)
	s.Register("test", hr, tr)

	_ = s.Observe("test", "livez", hr.Name, tr.Name)
	ob, _ := s.Observer("test", "livez")

	_ = s.Start(t.Context())

	testsubscriber.RequireObserverError(t, ob)
}

func TestValidObserver(t *testing.T) {
	s := server.NewServer()
	defer func() { _ = s.Stop(context.Background()) }()

	cc := checker.NewHTTPChecker(test.StatusURL("200"), timeout)
	hr := server.NewRegistration("http", period, cc)
	tc := checker.NewTCPChecker("google.com:80", timeout)
	tr := server.NewRegistration("tcp", period, tc)
	s.Register("test", hr, tr)

	_ = s.Observe("test", "livez", hr.Name, tr.Name)
	ob, _ := s.Observer("test", "livez")

	_ = s.Start(t.Context())

	testsubscriber.RequireObserverNoError(t, ob)
}

func TestOneInvalidObserver(t *testing.T) {
	s := server.NewServer()
	defer func() { _ = s.Stop(context.Background()) }()

	cc := checker.NewHTTPChecker(test.StatusURL("500"), timeout)
	hr := server.NewRegistration("http", period, cc)
	tc := checker.NewTCPChecker("google.com:80", timeout)
	tr := server.NewRegistration("tcp", period, tc)
	s.Register("test", hr, tr)

	_ = s.Observe("test", "livez", tr.Name)
	ob, _ := s.Observer("test", "livez")

	_ = s.Start(t.Context())

	testsubscriber.RequireObserverNoError(t, ob)
}

func TestObserveUnknownProbeNames(t *testing.T) {
	s := server.NewServer()
	defer func() { _ = s.Stop(context.Background()) }()

	cc := checker.NewHTTPChecker(test.StatusURL("200"), timeout)
	hr := server.NewRegistration("http", period, cc)
	tc := checker.NewTCPChecker("google.com:80", timeout)
	tr := server.NewRegistration("tcp", period, tc)
	s.Register("test", hr, tr)

	err := s.Observe("test", "livez", "http1", "tcp1")
	require.ErrorIs(t, err, server.ErrProbeNotFound)
	require.ErrorContains(t, err, "http1")
	require.ErrorContains(t, err, "tcp1")

	_, err = s.Observer("test", "livez")
	require.ErrorIs(t, err, server.ErrObserverNotFound)
}

func TestLivezObservers(t *testing.T) {
	s := server.NewServer()
	defer func() { _ = s.Stop(context.Background()) }()

	checker := checker.NewHTTPChecker(
		"https://www.google.com/",
		timeout,
		checker.WithRoundTripper(http.DefaultTransport),
	)
	r := server.NewRegistration("google", period, checker)
	s.Register("test", r)

	_ = s.Observe("test", "livez", r.Name)

	var names []string
	for name := range s.Observers("livez") {
		names = append(names, name)
		break
	}

	require.Equal(t, []string{"test"}, names)
}

func TestGRPCObservers(t *testing.T) {
	s := server.NewServer()
	defer func() { _ = s.Stop(context.Background()) }()

	checker := checker.NewHTTPChecker(
		"https://www.google.com/",
		timeout,
		checker.WithRoundTripper(http.DefaultTransport),
	)
	r := server.NewRegistration("google", period, checker)
	s.Register("test", r)

	_ = s.Observe("test", "livez", r.Name)

	var names []string
	for name := range s.Observers("grpc") {
		names = append(names, name)
		break
	}

	require.Empty(t, names)
}

func TestInvalidObservers(t *testing.T) {
	s := server.NewServer()
	defer func() { _ = s.Stop(context.Background()) }()

	checker := checker.NewHTTPChecker(
		"https://www.google.com/",
		timeout,
		checker.WithRoundTripper(http.DefaultTransport),
	)
	r := server.NewRegistration("google", period, checker)
	s.Register("test", r)
	require.Error(t, s.Observe("bob", "livez", r.Name))

	_ = s.Observe("test", "livez", r.Name)
	_, err := s.Observer("bob", "livez")
	require.Error(t, err)
}

func BenchmarkValidHTTPChecker(b *testing.B) {
	b.ReportAllocs()

	s := server.NewServer()
	defer func() { _ = s.Stop(context.Background()) }()

	checker := checker.NewHTTPChecker("https://www.google.com/", period)

	r := server.NewRegistration("google", period, checker)
	s.Register("test", r)

	_ = s.Observe("test", "livez", r.Name)
	ob, _ := s.Observer("test", "livez")

	_ = s.Start(b.Context())
	testsubscriber.WaitObserverNoError(b, ob)

	b.ResetTimer()

	for b.Loop() {
		if err := ob.Error(); err != nil {
			b.Fatal(err)
		}
	}
}
