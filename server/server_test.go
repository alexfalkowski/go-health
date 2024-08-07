package server_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alexfalkowski/go-health/checker"
	"github.com/alexfalkowski/go-health/server"
	. "github.com/smartystreets/goconvey/convey" //nolint:revive
)

const google = "google"

func defaultTimeout() time.Duration {
	return 2 * time.Second
}

func defaultPeriod() time.Duration {
	return 500 * time.Millisecond
}

func defaultWait() time.Duration {
	return 1 * time.Second
}

func TestDoubleStart(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()
		defer s.Stop()

		checker := checker.NewHTTPChecker("https://www.google.com/", http.DefaultTransport, defaultTimeout())
		r := server.NewRegistration(google, defaultPeriod(), checker)

		s.Register(r)

		ob := s.Observe(google)

		Convey("When I start the server", func() {
			s.Start()
			s.Start()

			time.Sleep(defaultWait())

			Convey("Then I should have no error from the observer", func() {
				So(ob.Error(), ShouldBeNil)
			})
		})
	})
}

func TestValidHTTPChecker(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()
		defer s.Stop()

		checker := checker.NewHTTPChecker("https://www.google.com/", http.DefaultTransport, 0)
		r := server.NewRegistration(google, defaultPeriod(), checker)

		s.Register(r)

		ob := s.Observe(google)

		Convey("When I start the server", func() {
			s.Start()

			time.Sleep(defaultWait())

			Convey("Then I should have no error from the observer", func() {
				time.Sleep(defaultWait())

				So(ob.Error(), ShouldBeNil)
			})
		})
	})
}

func TestInvalidURLHTTPChecker(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()
		defer s.Stop()

		name := "assaaasss"
		checker := checker.NewHTTPChecker("https://www.assaaasss.com/", http.DefaultTransport, defaultTimeout())
		r := server.NewRegistration(name, defaultPeriod(), checker)

		s.Register(r)

		ob := s.Observe(name)

		Convey("When I start the server", func() {
			s.Start()

			time.Sleep(defaultWait())

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

		name := "assaaasss"
		checker := checker.NewHTTPChecker(string([]byte{0x7f}), http.DefaultTransport, defaultTimeout())
		r := server.NewRegistration(name, defaultPeriod(), checker)

		s.Register(r)

		ob := s.Observe(name)

		Convey("When I start the server", func() {
			s.Start()

			time.Sleep(defaultWait())

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

		name := "http400"
		checker := checker.NewHTTPChecker("http://localhost:6000/v1/status/400", http.DefaultTransport, defaultTimeout())
		r := server.NewRegistration(name, defaultPeriod(), checker)

		s.Register(r)

		ob := s.Observe(name)

		Convey("When I start the server", func() {
			s.Start()

			time.Sleep(defaultWait())

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

		name := "http200"
		checker := checker.NewHTTPChecker("http://localhost:6000/v1/status/200?sleep=5s",
			http.DefaultTransport, defaultTimeout())
		r := server.NewRegistration(name, defaultPeriod(), checker)

		s.Register(r)

		ob := s.Observe(name)

		Convey("When I start the server", func() {
			s.Start()

			time.Sleep(defaultWait())

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

		name := "tcp-google"
		checker := checker.NewTCPChecker("www.google.com:80", defaultTimeout())
		r := server.NewRegistration(name, defaultPeriod(), checker)

		s.Register(r)

		ob := s.Observe(name)

		Convey("When I start the server", func() {
			s.Start()

			time.Sleep(defaultWait())

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

		name := "tcp-assaaasss"
		checker := checker.NewTCPChecker("www.assaaasss.com:80", defaultTimeout())
		r := server.NewRegistration(name, defaultPeriod(), checker)

		s.Register(r)

		ob := s.Observe(name)

		Convey("When I start the server", func() {
			s.Start()

			time.Sleep(defaultWait())

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

		name := "db"
		checker := checker.NewDBChecker(db, defaultTimeout())
		r := server.NewRegistration(name, defaultPeriod(), checker)

		s.Register(r)

		ob := s.Observe(name)

		Convey("When I start the server", func() {
			s.Start()

			time.Sleep(defaultWait())

			Convey("Then I should have no error from the observer", func() {
				So(ob.Error(), ShouldBeNil)
			})
		})
	})
}

func TestValidReadyChecker(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()
		defer s.Stop()

		name := "ready"
		errNotReady := errors.New("not ready")
		checker := checker.NewReadyChecker(errNotReady)
		r := server.NewRegistration(name, defaultPeriod(), checker)

		s.Register(r)

		ob := s.Observe(name)

		Convey("When I start the server", func() {
			s.Start()

			time.Sleep(defaultWait())

			Convey("Then I should have no error from the observer", func() {
				So(ob.Error(), ShouldEqual, errNotReady)

				checker.Ready()

				time.Sleep(defaultWait())

				So(ob.Error(), ShouldBeNil)
			})
		})
	})
}

func TestValidNoopChecker(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()
		defer s.Stop()

		name := "noop"
		checker := checker.NewNoopChecker()
		r := server.NewRegistration(name, defaultPeriod(), checker)

		s.Register(r)

		ob := s.Observe(name)

		Convey("When I start the server", func() {
			s.Start()

			time.Sleep(defaultWait())

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

		cc := checker.NewHTTPChecker("http://localhost:6000/v1/status/400", http.DefaultTransport, defaultTimeout())
		hr := server.NewRegistration("http1", defaultPeriod(), cc)
		tc := checker.NewTCPChecker("httpstat.us:9000", defaultTimeout())
		tr := server.NewRegistration("tcp1", defaultPeriod(), tc)

		s.Register(hr, tr)

		ob := s.Observe("http1", "tcp1")

		Convey("When I start the server", func() {
			s.Start()

			time.Sleep(defaultWait())

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

		cc := checker.NewHTTPChecker("http://localhost:6000/v1/status/200", http.DefaultTransport, defaultTimeout())
		hr := server.NewRegistration("http", defaultPeriod(), cc)
		tc := checker.NewTCPChecker("httpstat.us:80", defaultTimeout())
		tr := server.NewRegistration("tcp", defaultPeriod(), tc)

		s.Register(hr, tr)

		ob := s.Observe("http", "tcp")

		Convey("When I start the server", func() {
			s.Start()

			time.Sleep(defaultWait())

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

		cc := checker.NewHTTPChecker("http://localhost:6000/v1/status/500", http.DefaultTransport, defaultTimeout())
		hr := server.NewRegistration("http", defaultPeriod(), cc)
		tc := checker.NewTCPChecker("httpstat.us:80", defaultTimeout())
		tr := server.NewRegistration("tcp", defaultPeriod(), tc)

		s.Register(hr, tr)

		ob := s.Observe("tcp")

		Convey("When I start the server", func() {
			s.Start()

			time.Sleep(defaultWait())

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

		cc := checker.NewHTTPChecker("http://localhost:6000/v1/status/200", http.DefaultTransport, defaultTimeout())
		hr := server.NewRegistration("http", defaultPeriod(), cc)
		tc := checker.NewTCPChecker("httpstat.us:80", defaultTimeout())
		tr := server.NewRegistration("tcp", defaultPeriod(), tc)

		s.Register(hr, tr)

		Convey("When I observer a non existent registration", func() {
			ob := s.Observe("http1", "tcp1")

			Convey("Then I should have no error from the probe", func() {
				So(ob.Error(), ShouldBeNil)
			})
		})
	})
}

func BenchmarkValidHTTPChecker(b *testing.B) {
	b.ReportAllocs()

	s := server.NewServer()
	defer s.Stop()

	checker := checker.NewHTTPChecker("https://www.google.com/", http.DefaultTransport, defaultPeriod())

	r := server.NewRegistration(google, defaultPeriod(), checker)
	s.Register(r)

	ob := s.Observe(google)

	s.Start()
	time.Sleep(defaultWait())

	b.Run(google, func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if err := ob.Error(); err != nil {
				b.Fail()
			}
		}
	})
}
