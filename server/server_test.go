package server_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alexfalkowski/go-health/v2/checker"
	"github.com/alexfalkowski/go-health/v2/internal/test"
	"github.com/alexfalkowski/go-health/v2/net"
	"github.com/alexfalkowski/go-health/v2/server"
	"github.com/stretchr/testify/require"
)

const (
	timeout = 2 * time.Second
	period  = 500 * time.Millisecond
	wait    = 1 * time.Second
)

var invalidURL = string([]byte{0x7f})

func TestStop(t *testing.T) {
	s := server.NewServer()

	stopped := make(chan struct{})
	go func() {
		s.Stop()
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("server stop timed out")
	}
}

func TestDoubleStart(t *testing.T) {
	s := server.NewServer()
	defer s.Stop()

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

	s.Start()
	s.Start()
	time.Sleep(wait)

	require.NoError(t, ob.Error())
}

func TestOnlineChecker(t *testing.T) {
	s := server.NewServer()
	defer s.Stop()

	r := server.NewOnlineRegistration(0, period)
	s.Register("test", r)

	_ = s.Observe("test", "livez", r.Name)
	ob, _ := s.Observer("test", "livez")

	s.Start()
	time.Sleep(wait)

	require.NoError(t, ob.Error())
}

func TestInvalidOnlineChecker(t *testing.T) {
	s := server.NewServer()
	defer s.Stop()

	r := server.NewOnlineRegistration(0, period, checker.WithURLs(invalidURL, "https://www.assaaasss.com/"))
	s.Register("test", r)

	_ = s.Observe("test", "livez", r.Name)
	ob, _ := s.Observer("test", "livez")

	s.Start()
	time.Sleep(wait)

	require.Error(t, ob.Error())
}

func TestValidHTTPChecker(t *testing.T) {
	s := server.NewServer()
	defer s.Stop()

	checker := checker.NewHTTPChecker("https://www.google.com/", 0)
	r := server.NewRegistration("google", period, checker)
	s.Register("test", r)

	_ = s.Observe("test", "livez", r.Name)
	ob, _ := s.Observer("test", "livez")

	s.Start()
	time.Sleep(wait)

	require.NoError(t, ob.Error())
}

func TestInvalidURLHTTPChecker(t *testing.T) {
	s := server.NewServer()
	defer s.Stop()

	checker := checker.NewHTTPChecker("https://www.assaaasss.com/", timeout)
	r := server.NewRegistration("assaaasss", period, checker)
	s.Register("test", r)

	_ = s.Observe("test", "livez", r.Name)
	ob, _ := s.Observer("test", "livez")

	s.Start()
	time.Sleep(wait)

	require.Error(t, ob.Error())
}

func TestMalformedURLHTTPChecker(t *testing.T) {
	s := server.NewServer()
	defer s.Stop()

	checker := checker.NewHTTPChecker(invalidURL, timeout)
	r := server.NewRegistration("assaaasss", period, checker)
	s.Register("test", r)

	_ = s.Observe("test", "livez", r.Name)
	ob, _ := s.Observer("test", "livez")

	s.Start()
	time.Sleep(wait)

	require.Error(t, ob.Error())
}

func TestInvalidCodeHTTPChecker(t *testing.T) {
	s := server.NewServer()
	defer s.Stop()

	checker := checker.NewHTTPChecker(test.StatusURL("400"), timeout)
	r := server.NewRegistration("http400", period, checker)
	s.Register("test", r)

	_ = s.Observe("test", "livez", r.Name)
	ob, _ := s.Observer("test", "livez")

	s.Start()
	time.Sleep(wait)

	require.Error(t, ob.Error())
}

func TestTimeoutHTTPChecker(t *testing.T) {
	s := server.NewServer()
	defer s.Stop()

	checker := checker.NewHTTPChecker(test.StatusURL("200?sleep=5s"), timeout)
	r := server.NewRegistration("http200", period, checker)
	s.Register("test", r)

	_ = s.Observe("test", "livez", r.Name)
	ob, _ := s.Observer("test", "livez")

	s.Start()
	time.Sleep(wait)

	require.Error(t, ob.Error())
}

func TestValidTCPChecker(t *testing.T) {
	s := server.NewServer()
	defer s.Stop()

	checker := checker.NewTCPChecker(
		"www.google.com:80",
		timeout,
		checker.WithDialer(net.DefaultDialer),
	)
	r := server.NewRegistration("tcp-google", period, checker)
	s.Register("test", r)

	_ = s.Observe("test", "livez", r.Name)
	ob, _ := s.Observer("test", "livez")

	s.Start()
	time.Sleep(wait)

	require.NoError(t, ob.Error())
}

func TestInvalidAddressTCPChecker(t *testing.T) {
	s := server.NewServer()
	defer s.Stop()

	checker := checker.NewTCPChecker("www.assaaasss.com:80", timeout)
	r := server.NewRegistration("tcp-assaaasss", period, checker)
	s.Register("test", r)

	_ = s.Observe("test", "livez", r.Name)
	ob, _ := s.Observer("test", "livez")

	s.Start()
	time.Sleep(wait)

	require.Error(t, ob.Error())
	require.Error(t, ob.Errors()["tcp-assaaasss"])
}

func TestValidDBChecker(t *testing.T) {
	s := server.NewServer()
	defer s.Stop()

	db, _, err := sqlmock.New()
	require.NoError(t, err)

	defer db.Close()

	checker := checker.NewDBChecker(db, timeout)
	r := server.NewRegistration("db", period, checker)
	s.Register("test", r)

	_ = s.Observe("test", "livez", r.Name)
	ob, _ := s.Observer("test", "livez")

	s.Start()
	time.Sleep(wait)

	require.NoError(t, ob.Error())
}

//nolint:err113
func TestValidReadyChecker(t *testing.T) {
	s := server.NewServer()
	defer s.Stop()

	errNotReady := errors.New("not ready")
	checker := checker.NewReadyChecker(errNotReady)
	r := server.NewRegistration("ready", period, checker)
	s.Register("test", r)

	_ = s.Observe("test", "livez", r.Name)
	ob, _ := s.Observer("test", "livez")

	s.Start()
	time.Sleep(wait)

	require.Error(t, ob.Error())

	checker.Ready()
	time.Sleep(wait)

	require.NoError(t, ob.Error())
}

func TestValidNoopChecker(t *testing.T) {
	s := server.NewServer()
	defer s.Stop()

	checker := checker.NewNoopChecker()
	r := server.NewRegistration("noop", period, checker)
	s.Register("test", r)

	_ = s.Observe("test", "livez", r.Name)
	ob, _ := s.Observer("test", "livez")

	s.Start()
	time.Sleep(wait)

	require.NoError(t, ob.Error())
}

func TestInvalidObserver(t *testing.T) {
	s := server.NewServer()
	defer s.Stop()

	cc := checker.NewHTTPChecker(test.StatusURL("400"), timeout)
	hr := server.NewRegistration("http1", period, cc)
	tc := checker.NewTCPChecker("httpstat.us:9000", timeout)
	tr := server.NewRegistration("tcp1", period, tc)
	s.Register("test", hr, tr)

	_ = s.Observe("test", "livez", hr.Name, tr.Name)
	ob, _ := s.Observer("test", "livez")

	s.Start()
	time.Sleep(wait)

	require.Error(t, ob.Error())
}

func TestValidObserver(t *testing.T) {
	s := server.NewServer()
	defer s.Stop()

	cc := checker.NewHTTPChecker(test.StatusURL("200"), timeout)
	hr := server.NewRegistration("http", period, cc)
	tc := checker.NewTCPChecker("httpstat.us:80", timeout)
	tr := server.NewRegistration("tcp", period, tc)
	s.Register("test", hr, tr)

	_ = s.Observe("test", "livez", hr.Name, tr.Name)
	ob, _ := s.Observer("test", "livez")

	s.Start()
	time.Sleep(wait)

	require.NoError(t, ob.Error())
}

func TestOneInvalidObserver(t *testing.T) {
	s := server.NewServer()
	defer s.Stop()

	cc := checker.NewHTTPChecker(test.StatusURL("500"), timeout)
	hr := server.NewRegistration("http", period, cc)
	tc := checker.NewTCPChecker("httpstat.us:80", timeout)
	tr := server.NewRegistration("tcp", period, tc)
	s.Register("test", hr, tr)

	_ = s.Observe("test", "livez", tr.Name)
	ob, _ := s.Observer("test", "livez")

	s.Start()
	time.Sleep(wait)

	require.NoError(t, ob.Error())
}

func TestNonExistentObserver(t *testing.T) {
	s := server.NewServer()
	defer s.Stop()

	cc := checker.NewHTTPChecker(test.StatusURL("200"), timeout)
	hr := server.NewRegistration("http", period, cc)
	tc := checker.NewTCPChecker("httpstat.us:80", timeout)
	tr := server.NewRegistration("tcp", period, tc)
	s.Register("test", hr, tr)

	_ = s.Observe("test", "livez", "http1", "tcp1")
	ob, _ := s.Observer("test", "livez")

	require.NoError(t, ob.Error())
}

func TestLivezObservers(t *testing.T) {
	s := server.NewServer()
	defer s.Stop()

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
	defer s.Stop()

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
	defer s.Stop()

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
	defer s.Stop()

	checker := checker.NewHTTPChecker("https://www.google.com/", period)

	r := server.NewRegistration("google", period, checker)
	s.Register("test", r)

	_ = s.Observe("test", "livez", r.Name)
	ob, _ := s.Observer("test", "livez")

	s.Start()
	time.Sleep(wait)

	b.ResetTimer()

	for b.Loop() {
		require.Error(b, ob.Error())
	}
}
