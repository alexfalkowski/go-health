package server_test

import (
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alexfalkowski/go-health/pkg/checker"
	"github.com/alexfalkowski/go-health/pkg/server"
	. "github.com/smartystreets/goconvey/convey"
)

func defaultTimeout() time.Duration {
	return 500 * time.Millisecond
}

func defaultPeriod() time.Duration {
	return 500 * time.Millisecond
}

func defaultWait() time.Duration {
	return 2 * time.Second
}

func TestNoRegistrations(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()

		Convey("When I start the server", func() {
			err := s.Start()

			Convey("Then I should have an error", func() {
				So(err, ShouldBeError)
			})
		})

		Convey("When I stop the server", func() {
			err := s.Stop()

			Convey("Then I should have an error", func() {
				So(err, ShouldBeError)
			})
		})
	})
}

func TestDuplicateRegistrations(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()
		checker := checker.NewHTTPChecker("https://www.google.com/", defaultTimeout())
		r := server.NewRegistration("google", defaultPeriod(), checker)

		_ = s.Register(r)

		Convey("When we add a duplicate subscriber", func() {
			err := s.Register(r)

			Convey("Then I should have an error", func() {
				So(err, ShouldBeError)
			})
		})
	})
}

func TestNonExistentRegistration(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()
		defer s.Stop() // nolint:errcheck

		checker := checker.NewHTTPChecker("https://www.google.com/", defaultTimeout())
		r := server.NewRegistration("google", defaultPeriod(), checker)

		_ = s.Register(r)

		Convey("When I subscribe to non existent registration", func() {
			_, err := s.Subscribe("google1")

			Convey("Then I should have an error", func() {
				So(err, ShouldBeError)
			})
		})
	})
}

func TestDoubleStart(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()
		defer s.Stop() // nolint:errcheck

		name := "google"
		checker := checker.NewHTTPChecker("https://www.google.com/", defaultTimeout())
		r := server.NewRegistration(name, defaultPeriod(), checker)

		_ = s.Register(r)

		ob, _ := s.Observe(name)

		Convey("When I start the server", func() {
			err := s.Start()
			So(err, ShouldBeNil)

			err = s.Start()
			So(err, ShouldBeNil)

			Convey("Then I should have no error from the observer", func() {
				time.Sleep(defaultWait())

				So(ob.Error(), ShouldBeNil)
			})
		})
	})
}

func TestValidHTTPChecker(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()
		defer s.Stop() // nolint:errcheck

		name := "google"
		checker := checker.NewHTTPChecker("https://www.google.com/", defaultTimeout())
		r := server.NewRegistration(name, defaultPeriod(), checker)

		_ = s.Register(r)

		ob, _ := s.Observe(name)

		Convey("When I start the server", func() {
			err := s.Start()
			So(err, ShouldBeNil)

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
		defer s.Stop() // nolint:errcheck

		name := "assaaasss"
		checker := checker.NewHTTPChecker("https://www.assaaasss.com/", defaultTimeout())
		r := server.NewRegistration(name, defaultPeriod(), checker)

		_ = s.Register(r)

		ob, _ := s.Observe(name)

		Convey("When I start the server", func() {
			err := s.Start()
			So(err, ShouldBeNil)

			Convey("Then I should have error from the observer", func() {
				So(ob.Error(), ShouldBeError)
			})
		})
	})
}

func TestInvalidCodeHTTPChecker(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()
		defer s.Stop() // nolint:errcheck

		name := "httpstat400"
		checker := checker.NewHTTPChecker("https://httpstat.us/400", defaultTimeout())
		r := server.NewRegistration(name, defaultPeriod(), checker)

		_ = s.Register(r)

		ob, _ := s.Observe(name)

		Convey("When I start the server", func() {
			err := s.Start()
			So(err, ShouldBeNil)

			Convey("Then I should have error from the observer", func() {
				So(ob.Error(), ShouldBeError)
			})
		})
	})
}

func TestTimeoutHTTPChecker(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()
		defer s.Stop() // nolint:errcheck

		name := "httpstat200"
		checker := checker.NewHTTPChecker("https://httpstat.us/200?sleep=6000", defaultTimeout())
		r := server.NewRegistration(name, defaultPeriod(), checker)

		_ = s.Register(r)

		ob, _ := s.Observe(name)

		Convey("When I start the server", func() {
			err := s.Start()
			So(err, ShouldBeNil)

			Convey("Then I should have error from the observer", func() {
				So(ob.Error(), ShouldBeError)
			})
		})
	})
}

func TestValidTCPChecker(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()
		defer s.Stop() // nolint:errcheck

		name := "tcp-google"
		checker := checker.NewTCPChecker("www.google.com:80", defaultTimeout())
		r := server.NewRegistration(name, defaultPeriod(), checker)

		_ = s.Register(r)

		ob, _ := s.Observe(name)

		Convey("When I start the server", func() {
			err := s.Start()
			So(err, ShouldBeNil)

			Convey("Then I should have no error from the observer", func() {
				So(ob.Error(), ShouldBeNil)
			})
		})
	})
}

func TestInvalidAddressTCPChecker(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()
		defer s.Stop() // nolint:errcheck

		name := "tcp-assaaasss"
		checker := checker.NewTCPChecker("www.assaaasss.com:80", defaultTimeout())
		r := server.NewRegistration(name, defaultPeriod(), checker)

		_ = s.Register(r)

		ob, _ := s.Observe(name)

		Convey("When I start the server", func() {
			err := s.Start()
			So(err, ShouldBeNil)

			Convey("Then I should have error from the observer", func() {
				So(ob.Error(), ShouldBeError)
			})
		})
	})
}

func TestValidDBChecker(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()
		defer s.Stop() // nolint:errcheck

		db, _, err := sqlmock.New()
		So(err, ShouldBeNil)

		defer db.Close()

		name := "db"
		checker := checker.NewDBChecker(db, defaultTimeout())
		r := server.NewRegistration(name, defaultPeriod(), checker)

		_ = s.Register(r)

		ob, _ := s.Observe(name)

		Convey("When I start the server", func() {
			err := s.Start()
			So(err, ShouldBeNil)

			Convey("Then I should have no error from the observer", func() {
				So(ob.Error(), ShouldBeNil)
			})
		})
	})
}

func TestValidReadyChecker(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()
		defer s.Stop() // nolint:errcheck

		name := "ready"
		errNotReady := errors.New("not ready")
		checker := checker.NewReadyChecker(errNotReady)
		r := server.NewRegistration(name, defaultPeriod(), checker)

		_ = s.Register(r)

		ob, _ := s.Observe(name)

		Convey("When I start the server", func() {
			err := s.Start()
			So(err, ShouldBeNil)

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
		defer s.Stop() // nolint:errcheck

		name := "noop"
		checker := checker.NewNoopChecker()
		r := server.NewRegistration(name, defaultPeriod(), checker)

		_ = s.Register(r)

		ob, _ := s.Observe(name)

		Convey("When I start the server", func() {
			err := s.Start()
			So(err, ShouldBeNil)

			Convey("Then I should have no error from the observer", func() {
				So(ob.Error(), ShouldBeNil)
			})
		})
	})
}

func TestInvalidObserver(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()
		defer s.Stop() // nolint:errcheck

		cc := checker.NewHTTPChecker("https://httpstat.us/400", defaultTimeout())
		hr := server.NewRegistration("http1", defaultPeriod(), cc)
		tc := checker.NewTCPChecker("httpstat.us:9000", defaultTimeout())
		tr := server.NewRegistration("tcp1", defaultPeriod(), tc)

		_ = s.Register(hr, tr)

		ob, _ := s.Observe("http1", "tcp1")

		Convey("When I start the server", func() {
			err := s.Start()
			So(err, ShouldBeNil)

			Convey("Then I should have error from the probe", func() {
				So(ob.Error(), ShouldBeError)
			})
		})
	})
}

func TestValidObserver(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()
		defer s.Stop() // nolint:errcheck

		cc := checker.NewHTTPChecker("https://httpstat.us/200", defaultTimeout())
		hr := server.NewRegistration("http", defaultPeriod(), cc)
		tc := checker.NewTCPChecker("httpstat.us:80", defaultTimeout())
		tr := server.NewRegistration("tcp", defaultPeriod(), tc)

		_ = s.Register(hr, tr)

		ob, _ := s.Observe("http", "tcp")

		Convey("When I start the server", func() {
			err := s.Start()
			So(err, ShouldBeNil)

			Convey("Then I should have no error from the probe", func() {
				So(ob.Error(), ShouldBeNil)
			})
		})
	})
}

func TestNonExistentObserver(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		s := server.NewServer()
		defer s.Stop() // nolint:errcheck

		cc := checker.NewHTTPChecker("https://httpstat.us/200", defaultTimeout())
		hr := server.NewRegistration("http", defaultPeriod(), cc)
		tc := checker.NewTCPChecker("httpstat.us:80", defaultTimeout())
		tr := server.NewRegistration("tcp", defaultPeriod(), tc)

		_ = s.Register(hr, tr)

		Convey("When I observer a non existent registration", func() {
			_, err := s.Observe("http1", "tcp1")

			Convey("Then I should have an error", func() {
				So(err, ShouldBeError)
			})
		})
	})
}
