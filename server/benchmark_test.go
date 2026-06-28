package server_test

import (
	"context"
	"testing"

	"github.com/alexfalkowski/go-health/v2/checker"
	testsubscriber "github.com/alexfalkowski/go-health/v2/internal/test/subscriber"
	"github.com/alexfalkowski/go-health/v2/server"
)

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
