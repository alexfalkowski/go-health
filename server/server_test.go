package server_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alexfalkowski/go-health/checker"
	"github.com/alexfalkowski/go-health/server"
	. "github.com/smartystreets/goconvey/convey"
)

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

		name := "google"
		checker := checker.NewHTTPChecker("https://www.google.com/", &http.Client{Timeout: defaultTimeout()})
		r := server.NewRegistration(name, defaultPeriod(), checker)

		s.Register(r)

		ob := s.Observe(name)

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

		name := "google"
		checker := checker.NewHTTPChecker("https://www.google.com/", &http.Client{Timeout: defaultTimeout()})
		r := server.NewRegistration(name, defaultPeriod(), checker)

		s.Register(r)

		ob := s.Observe(name)

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
		checker := checker.NewHTTPChecker("https://www.assaaasss.com/", &http.Client{Timeout: defaultTimeout()})
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

func TestMallformedURLHTTPChecker(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()
		defer s.Stop()

		name := "assaaasss"
		checker := checker.NewHTTPChecker(string([]byte{0x7f}), &http.Client{Timeout: defaultTimeout()})
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

		name := "httpstat400"
		checker := checker.NewHTTPChecker("https://httpstat.us/400", &http.Client{Timeout: defaultTimeout()})
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

		name := "httpstat200"
		checker := checker.NewHTTPChecker("https://httpstat.us/200?sleep=6000", &http.Client{Timeout: defaultTimeout()})
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

// nolint:goerr113
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

		cc := checker.NewHTTPChecker("https://httpstat.us/400", &http.Client{Timeout: defaultTimeout()})
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

		cc := checker.NewHTTPChecker("https://httpstat.us/200", &http.Client{Timeout: defaultTimeout()})
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

		cc := checker.NewHTTPChecker("https://httpstat.us/500", &http.Client{Timeout: defaultTimeout()})
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

		cc := checker.NewHTTPChecker("https://httpstat.us/200", &http.Client{Timeout: defaultTimeout()})
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
