package svr_test

import (
	"testing"
	"time"

	"github.com/alexfalkowski/go-health/pkg/chk"
	"github.com/alexfalkowski/go-health/pkg/svr"
	. "github.com/smartystreets/goconvey/convey"
)

func defaultTimeout() time.Duration {
	return 10 * time.Second
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
		checker := chk.NewHTTPChecker("https://www.google.com/", defaultTimeout(), nil)

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
		checker := chk.NewHTTPChecker("https://www.google.com/", defaultTimeout(), nil)

		_ = server.Register(name, defaultPeriod(), checker)

		sub := server.Subscribe(name)

		Convey("When I start the server", func() {
			err := server.Start()

			Convey("Then I should have no server error", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then I should have no error from the probe", func() {
				err := <-sub.Receive()
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestInvalidURLHTTPChecker(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		server := svr.NewServer()
		defer server.Stop() // nolint:errcheck

		name := "assaaasss"
		checker := chk.NewHTTPChecker("https://www.assaaasss.com/", defaultTimeout(), nil)

		_ = server.Register(name, defaultPeriod(), checker)

		sub := server.Subscribe(name)

		Convey("When I start the server", func() {
			err := server.Start()

			Convey("Then I should have no server error", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then I should have error from the probe", func() {
				err := <-sub.Receive()
				So(err, ShouldBeError)
			})
		})
	})
}

func TestInvalidCodeHTTPChecker(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		server := svr.NewServer()
		defer server.Stop() // nolint:errcheck

		name := "httpstat"
		checker := chk.NewHTTPChecker("https://httpstat.us/400", defaultTimeout(), nil)

		_ = server.Register(name, defaultPeriod(), checker)

		sub := server.Subscribe(name)

		Convey("When I start the server", func() {
			err := server.Start()

			Convey("Then I should have no server error", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then I should have error from the probe", func() {
				err := <-sub.Receive()
				So(err, ShouldBeError)
			})
		})
	})
}

func TestTimeoutHTTPChecker(t *testing.T) {
	Convey("Given we have a new server", t, func() {
		server := svr.NewServer()
		defer server.Stop() // nolint:errcheck

		name := "httpstat"
		checker := chk.NewHTTPChecker("https://httpstat.us/200?sleep=20000", defaultTimeout(), nil)

		_ = server.Register(name, defaultPeriod(), checker)

		sub := server.Subscribe(name)

		Convey("When I start the server", func() {
			err := server.Start()

			Convey("Then I should have no server error", func() {
				So(err, ShouldBeNil)
			})

			Convey("Then I should have error from the probe", func() {
				err := <-sub.Receive()
				So(err, ShouldBeError)
			})
		})
	})
}
