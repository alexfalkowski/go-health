package server_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alexfalkowski/go-health/v2/checker"
	"github.com/alexfalkowski/go-health/v2/net"
	"github.com/alexfalkowski/go-health/v2/server"
	. "github.com/smartystreets/goconvey/convey" //nolint:revive
)

const (
	timeout = 2 * time.Second
	period  = 500 * time.Millisecond
	wait    = 1 * time.Second
)

func TestDoubleStart(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()
		defer s.Stop()

		checker := checker.NewHTTPChecker(
			"https://www.google.com/",
			timeout,
			checker.WithRoundTripper(http.DefaultTransport),
		)
		r := server.NewRegistration("google", period, checker)
		s.Register("test", r)

		Convey("When I start the server", func() {
			_ = s.Observe("test", "livez", r.Name)
			_, _ = s.Observer("test", "livez")

			_ = s.Observe("test", "livez", r.Name)
			ob, _ := s.Observer("test", "livez")

			s.Start()
			s.Start()
			time.Sleep(wait)

			Convey("Then I should have no error from the observer", func() {
				So(ob.Error(), ShouldBeNil)
			})
		})
	})
}

func TestOnlineChecker(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()
		defer s.Stop()

		r := server.NewOnlineRegistration(0, period)
		s.Register("test", r)

		Convey("When I start the server", func() {
			_ = s.Observe("test", "livez", r.Name)
			ob, _ := s.Observer("test", "livez")

			s.Start()
			time.Sleep(wait)

			Convey("Then I should have no error from the observer", func() {
				time.Sleep(wait)

				So(ob.Error(), ShouldBeNil)
			})
		})
	})
}

func TestValidHTTPChecker(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()
		defer s.Stop()

		checker := checker.NewHTTPChecker("https://www.google.com/", 0)
		r := server.NewRegistration("google", period, checker)
		s.Register("test", r)

		Convey("When I start the server", func() {
			_ = s.Observe("test", "livez", r.Name)
			ob, _ := s.Observer("test", "livez")

			s.Start()
			time.Sleep(wait)

			Convey("Then I should have no error from the observer", func() {
				time.Sleep(wait)

				So(ob.Error(), ShouldBeNil)
			})
		})
	})
}

func TestInvalidURLHTTPChecker(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()
		defer s.Stop()

		checker := checker.NewHTTPChecker("https://www.assaaasss.com/", timeout)
		r := server.NewRegistration("assaaasss", period, checker)
		s.Register("test", r)

		Convey("When I start the server", func() {
			_ = s.Observe("test", "livez", r.Name)
			ob, _ := s.Observer("test", "livez")

			s.Start()
			time.Sleep(wait)

			Convey("Then I should have error from the observer", func() {
				So(ob.Error(), ShouldBeError)
			})
		})
	})
}

func TestMalformedURLHTTPChecker(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()
		defer s.Stop()

		checker := checker.NewHTTPChecker(string([]byte{0x7f}), timeout)
		r := server.NewRegistration("assaaasss", period, checker)
		s.Register("test", r)

		Convey("When I start the server", func() {
			_ = s.Observe("test", "livez", r.Name)
			ob, _ := s.Observer("test", "livez")

			s.Start()
			time.Sleep(wait)

			Convey("Then I should have error from the observer", func() {
				So(ob.Error(), ShouldBeError)
			})
		})
	})
}

func TestInvalidCodeHTTPChecker(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()
		defer s.Stop()

		checker := checker.NewHTTPChecker("http://localhost:6000/v1/status/400", timeout)
		r := server.NewRegistration("http400", period, checker)
		s.Register("test", r)

		Convey("When I start the server", func() {
			_ = s.Observe("test", "livez", r.Name)
			ob, _ := s.Observer("test", "livez")

			s.Start()
			time.Sleep(wait)

			Convey("Then I should have error from the observer", func() {
				So(ob.Error(), ShouldBeError)
			})
		})
	})
}

func TestTimeoutHTTPChecker(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()
		defer s.Stop()

		checker := checker.NewHTTPChecker("http://localhost:6000/v1/status/200?sleep=5s", timeout)
		r := server.NewRegistration("http200", period, checker)
		s.Register("test", r)

		Convey("When I start the server", func() {
			_ = s.Observe("test", "livez", r.Name)
			ob, _ := s.Observer("test", "livez")

			s.Start()
			time.Sleep(wait)

			Convey("Then I should have error from the observer", func() {
				So(ob.Error(), ShouldBeError)
			})
		})
	})
}

func TestValidTCPChecker(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()
		defer s.Stop()

		checker := checker.NewTCPChecker(
			"www.google.com:80",
			timeout,
			checker.WithDialer(net.DefaultDialer),
		)
		r := server.NewRegistration("tcp-google", period, checker)
		s.Register("test", r)

		Convey("When I start the server", func() {
			_ = s.Observe("test", "livez", r.Name)
			ob, _ := s.Observer("test", "livez")

			s.Start()
			time.Sleep(wait)

			Convey("Then I should have no error from the observer", func() {
				So(ob.Error(), ShouldBeNil)
			})
		})
	})
}

func TestInvalidAddressTCPChecker(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()
		defer s.Stop()

		checker := checker.NewTCPChecker("www.assaaasss.com:80", timeout)
		r := server.NewRegistration("tcp-assaaasss", period, checker)
		s.Register("test", r)

		Convey("When I start the server", func() {
			_ = s.Observe("test", "livez", r.Name)
			ob, _ := s.Observer("test", "livez")

			s.Start()
			time.Sleep(wait)

			Convey("Then I should have error from the observer", func() {
				So(ob.Error(), ShouldBeError)
				So(ob.Errors()["tcp-assaaasss"], ShouldBeError)
			})
		})
	})
}

func TestValidDBChecker(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()
		defer s.Stop()

		db, _, err := sqlmock.New()
		So(err, ShouldBeNil)

		defer db.Close()

		checker := checker.NewDBChecker(db, timeout)
		r := server.NewRegistration("db", period, checker)
		s.Register("test", r)

		Convey("When I start the server", func() {
			_ = s.Observe("test", "livez", r.Name)
			ob, _ := s.Observer("test", "livez")

			s.Start()
			time.Sleep(wait)

			Convey("Then I should have no error from the observer", func() {
				So(ob.Error(), ShouldBeNil)
			})
		})
	})
}

//nolint:err113
func TestValidReadyChecker(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()
		defer s.Stop()

		errNotReady := errors.New("not ready")
		checker := checker.NewReadyChecker(errNotReady)
		r := server.NewRegistration("ready", period, checker)
		s.Register("test", r)

		Convey("When I start the server", func() {
			_ = s.Observe("test", "livez", r.Name)
			ob, _ := s.Observer("test", "livez")

			s.Start()
			time.Sleep(wait)

			Convey("Then I should have no error from the observer", func() {
				So(ob.Error(), ShouldBeError)

				checker.Ready()
				time.Sleep(wait)
				So(ob.Error(), ShouldBeNil)
			})
		})
	})
}

func TestValidNoopChecker(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()
		defer s.Stop()

		checker := checker.NewNoopChecker()
		r := server.NewRegistration("noop", period, checker)
		s.Register("test", r)

		Convey("When I start the server", func() {
			_ = s.Observe("test", "livez", r.Name)
			ob, _ := s.Observer("test", "livez")

			s.Start()
			time.Sleep(wait)

			Convey("Then I should have no error from the observer", func() {
				So(ob.Error(), ShouldBeNil)
			})
		})
	})
}

func TestInvalidObserver(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()
		defer s.Stop()

		cc := checker.NewHTTPChecker("http://localhost:6000/v1/status/400", timeout)
		hr := server.NewRegistration("http1", period, cc)
		tc := checker.NewTCPChecker("httpstat.us:9000", timeout)
		tr := server.NewRegistration("tcp1", period, tc)
		s.Register("test", hr, tr)

		Convey("When I start the server", func() {
			_ = s.Observe("test", "livez", hr.Name, tr.Name)
			ob, _ := s.Observer("test", "livez")

			s.Start()
			time.Sleep(wait)

			Convey("Then I should have error from the probe", func() {
				So(ob.Error(), ShouldBeError)
			})
		})
	})
}

func TestValidObserver(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()
		defer s.Stop()

		cc := checker.NewHTTPChecker("http://localhost:6000/v1/status/200", timeout)
		hr := server.NewRegistration("http", period, cc)
		tc := checker.NewTCPChecker("httpstat.us:80", timeout)
		tr := server.NewRegistration("tcp", period, tc)
		s.Register("test", hr, tr)

		Convey("When I start the server", func() {
			_ = s.Observe("test", "livez", hr.Name, tr.Name)
			ob, _ := s.Observer("test", "livez")

			s.Start()
			time.Sleep(wait)

			Convey("Then I should have no error from the probe", func() {
				So(ob.Error(), ShouldBeNil)
			})
		})
	})
}

func TestOneInvalidObserver(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()
		defer s.Stop()

		cc := checker.NewHTTPChecker("http://localhost:6000/v1/status/500", timeout)
		hr := server.NewRegistration("http", period, cc)
		tc := checker.NewTCPChecker("httpstat.us:80", timeout)
		tr := server.NewRegistration("tcp", period, tc)
		s.Register("test", hr, tr)

		Convey("When I start the server", func() {
			_ = s.Observe("test", "livez", tr.Name)
			ob, _ := s.Observer("test", "livez")

			s.Start()
			time.Sleep(wait)

			Convey("Then I should have no error from the probe", func() {
				So(ob.Error(), ShouldBeNil)
			})
		})
	})
}

func TestNonExistentObserver(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()
		defer s.Stop()

		cc := checker.NewHTTPChecker("http://localhost:6000/v1/status/200", timeout)
		hr := server.NewRegistration("http", period, cc)
		tc := checker.NewTCPChecker("httpstat.us:80", timeout)
		tr := server.NewRegistration("tcp", period, tc)
		s.Register("test", hr, tr)

		Convey("When I observer a non existent registration", func() {
			_ = s.Observe("test", "livez", "http1", "tcp1")
			ob, _ := s.Observer("test", "livez")

			Convey("Then I should have no error from the probe", func() {
				So(ob.Error(), ShouldBeNil)
			})
		})
	})
}

func TestObservers(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()

		checker := checker.NewHTTPChecker(
			"https://www.google.com/",
			timeout,
			checker.WithRoundTripper(http.DefaultTransport),
		)
		r := server.NewRegistration("google", period, checker)
		s.Register("test", r)

		Convey("When I get livez observers", func() {
			_ = s.Observe("test", "livez", r.Name)

			var names []string
			for name := range s.Observers("livez") {
				names = append(names, name)
				break
			}

			Convey("Then I should have an observer", func() {
				So(names, ShouldEqual, []string{"test"})
			})
		})

		Convey("When I get grpc observers", func() {
			_ = s.Observe("test", "livez", r.Name)

			var names []string
			for name := range s.Observers("grpc") {
				names = append(names, name)
			}

			Convey("Then I should have no observers", func() {
				So(names, ShouldBeEmpty)
			})
		})
	})
}

func TestInvalidObservers(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()

		checker := checker.NewHTTPChecker(
			"https://www.google.com/",
			timeout,
			checker.WithRoundTripper(http.DefaultTransport),
		)
		r := server.NewRegistration("google", period, checker)
		s.Register("test", r)

		Convey("When I observe for a missing service", func() {
			err := s.Observe("bob", "livez", r.Name)

			Convey("Then I should have an error", func() {
				So(err, ShouldBeError)
			})
		})

		Convey("When I get an observer for a missing service", func() {
			_ = s.Observe("test", "livez", r.Name)
			_, err := s.Observer("bob", "livez")

			Convey("Then I should have an error", func() {
				So(err, ShouldBeError)
			})
		})
	})
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
		if err := ob.Error(); err != nil {
			b.Fail()
		}
	}
}
