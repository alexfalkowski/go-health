package subscriber_test

import (
	"errors"
	"testing"

	"github.com/alexfalkowski/go-health/v2/subscriber"
	"github.com/stretchr/testify/require"
)

func FuzzErrorsAggregationAndCopy(f *testing.F) {
	// Fuzz names and failure states because Errors owns named aggregation and defensive map copies.
	f.Add("db", true, "cache", true, "search", false)
	f.Add("", true, "cache", false, "search", true)
	f.Add("db:primary", true, "cache\nprimary", true, "search primary", true)
	f.Add("duplicate", true, "duplicate", false, "duplicate", true)

	f.Fuzz(func(t *testing.T, name1 string, failed1 bool, name2 string, failed2 bool, name3 string, failed3 bool) {
		errs := subscriber.Errors{
			name1: errorFor("first", failed1),
			name2: errorFor("second", failed2),
			name3: errorFor("third", failed3),
		}
		want := map[string]error{
			name1: errs[name1],
			name2: errs[name2],
			name3: errs[name3],
		}

		err := errs.Error()
		for name, wantErr := range want {
			if wantErr == nil {
				continue
			}

			require.ErrorIs(t, err, wantErr)
			require.ErrorContains(t, err, name+":")
		}
		if noFailures(want) {
			require.NoError(t, err)
		}

		copied := errs.Errors()
		copied[name1] = nil
		delete(copied, name2)

		for name, wantErr := range want {
			require.ErrorIs(t, errs[name], wantErr)
		}
	})
}

func errorFor(name string, failed bool) error {
	if failed {
		return errors.New(name + " failed")
	}

	return nil
}

func noFailures(errs map[string]error) bool {
	for _, err := range errs {
		if err != nil {
			return false
		}
	}

	return true
}
