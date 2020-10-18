package svr_test

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/alexfalkowski/go-health/pkg/chk"
	"github.com/alexfalkowski/go-health/pkg/svr"
	. "github.com/smartystreets/goconvey/convey"
)

func defaultTimeout() time.Duration {
	return 5 * time.Second
}

func defaultPeriod() time.Duration {
	return 1 * time.Second
}

func TestNoRegistrations(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		server := svr.NewServer()

		Convey("When I start the server", func() {
			err := server.Start()

			Convey("Then I should have an error", func() {
				So(err, ShouldBeError)
			})
		})

		Convey("When I stop the server", func() {
			err := server.Stop()

			Convey("Then I should have an error", func() {
				So(err, ShouldBeError)
			})
		})
	})
}

func TestDuplicateRegistrations(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		server := svr.NewServer()
		name := "google"
		checker := chk.NewHTTPChecker("https://www.google.com/", defaultTimeout())

		_ = server.Register(name, defaultPeriod(), checker)

		Convey("When add a duplicate subscriber", func() {
			err := server.Register(name, defaultPeriod(), checker)

			Convey("Then I should have an error", func() {
				So(err, ShouldBeError)
			})
		})
	})
}

func TestValidHTTPChecker(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		server := svr.NewServer()
		defer server.Stop() // nolint:errcheck

		name := "google"
		checker := chk.NewHTTPChecker("https://www.google.com/", defaultTimeout())

		_ = server.Register(name, defaultPeriod(), checker)

		sub := server.Subscribe(name)

		Convey("When I start the server", func() {
			err := server.Start()

			Convey("Then I should have no server error", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then I should have no error from the probe", func() {
				t := <-sub.Receive()
				So(t.Error(), ShouldBeNil)
			})
		})
	})
}

func TestInvalidURLHTTPChecker(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		server := svr.NewServer()
		defer server.Stop() // nolint:errcheck

		name := "assaaasss"
		checker := chk.NewHTTPChecker("https://www.assaaasss.com/", defaultTimeout())

		_ = server.Register(name, defaultPeriod(), checker)

		sub := server.Subscribe(name)

		Convey("When I start the server", func() {
			err := server.Start()

			Convey("Then I should have no server error", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then I should have error from the probe", func() {
				t := <-sub.Receive()
				So(t.Error(), ShouldBeError)
			})
		})
	})
}

func TestInvalidCodeHTTPChecker(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		server := svr.NewServer()
		defer server.Stop() // nolint:errcheck

		name := "httpstat400"
		checker := chk.NewHTTPChecker("https://httpstat.us/400", defaultTimeout())

		_ = server.Register(name, defaultPeriod(), checker)

		sub := server.Subscribe(name)

		Convey("When I start the server", func() {
			err := server.Start()

			Convey("Then I should have no server error", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then I should have error from the probe", func() {
				t := <-sub.Receive()
				So(t.Error(), ShouldBeError)
			})
		})
	})
}

func TestTimeoutHTTPChecker(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		server := svr.NewServer()
		defer server.Stop() // nolint:errcheck

		name := "httpstat200"
		checker := chk.NewHTTPChecker("https://httpstat.us/200?sleep=6000", defaultTimeout())

		_ = server.Register(name, defaultPeriod(), checker)

		sub := server.Subscribe(name)

		Convey("When I start the server", func() {
			err := server.Start()

			Convey("Then I should have no server error", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then I should have error from the probe", func() {
				t := <-sub.Receive()
				So(t.Error(), ShouldBeError)
			})
		})
	})
}

func TestValidTCPChecker(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		server := svr.NewServer()
		defer server.Stop() // nolint:errcheck

		name := "tcp-google"
		checker := chk.NewTCPChecker("www.google.com:80", defaultTimeout())

		_ = server.Register(name, defaultPeriod(), checker)

		sub := server.Subscribe(name)

		Convey("When I start the server", func() {
			err := server.Start()

			Convey("Then I should have no server error", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then I should have no error from the probe", func() {
				t := <-sub.Receive()
				So(t.Error(), ShouldBeNil)
			})
		})
	})
}

func TestInvalidAddressTCPChecker(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		server := svr.NewServer()
		defer server.Stop() // nolint:errcheck

		name := "tcp-assaaasss"
		checker := chk.NewTCPChecker("www.assaaasss.com:80", defaultTimeout())

		_ = server.Register(name, defaultPeriod(), checker)

		sub := server.Subscribe(name)

		Convey("When I start the server", func() {
			err := server.Start()

			Convey("Then I should have no server error", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then I should have error from the probe", func() {
				t := <-sub.Receive()
				So(t.Error(), ShouldBeError)
			})
		})
	})
}

func TestValidDBChecker(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		server := svr.NewServer()
		defer server.Stop() // nolint:errcheck

		db, _, err := sqlmock.New()
		So(err, ShouldBeNil)

		defer db.Close()

		name := "db"
		checker := chk.NewDBChecker(db, defaultTimeout())

		_ = server.Register(name, defaultPeriod(), checker)

		sub := server.Subscribe(name)

		Convey("When I start the server", func() {
			err := server.Start()

			Convey("Then I should have no server error", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then I should have no error from the probe", func() {
				t := <-sub.Receive()
				So(t.Error(), ShouldBeNil)
			})
		})
	})
}

func TestInvalidObserver(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		server := svr.NewServer()
		defer server.Stop() // nolint:errcheck

		_ = server.Register("http1", defaultPeriod(), chk.NewHTTPChecker("https://httpstat.us/400", defaultTimeout()))
		_ = server.Register("tcp1", defaultPeriod(), chk.NewTCPChecker("httpstat.us:9000", defaultTimeout()))

		ob := server.Observe("http1", "tcp1")

		Convey("When I start the server", func() {
			err := server.Start()

			Convey("Then I should have no server error", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then I should have error from the probe", func() {
				// Sleep for a period to make sure we get a result.
				time.Sleep(1750 * time.Millisecond)

				So(ob.Error(), ShouldBeError)
			})
		})
	})
}

func TestValidObserver(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		server := svr.NewServer()
		defer server.Stop() // nolint:errcheck

		_ = server.Register("http", defaultPeriod(), chk.NewHTTPChecker("https://httpstat.us/200", defaultTimeout()))
		_ = server.Register("tcp", defaultPeriod(), chk.NewTCPChecker("httpstat.us:80", defaultTimeout()))

		ob := server.Observe("http", "tcp")

		Convey("When I start the server", func() {
			err := server.Start()

			Convey("Then I should have no server error", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then I should have no error from the probe", func() {
				// Sleep for a period to make sure we get a result.
				time.Sleep(1750 * time.Millisecond)

				So(ob.Error(), ShouldBeNil)
			})
		})
	})
}
