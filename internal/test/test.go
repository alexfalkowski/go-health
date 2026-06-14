// Package test contains internal test helpers.
package test

import (
	"cmp"
	"fmt"
	"os"
)

// StatusURL returns the URL for the local status test service.
//
// It reads STATUS_PORT and falls back to 6000, producing URLs in the form
// http://localhost:<port>/v1/status/<status>.
func StatusURL(status string) string {
	port := cmp.Or(os.Getenv("STATUS_PORT"), "6000")
	return fmt.Sprintf("http://localhost:%s/v1/status/%s", port, status)
}

// ChannelClosed reports whether ch is closed.
func ChannelClosed(ch <-chan struct{}) bool {
	select {
	case <-ch:
		return true
	default:
		return false
	}
}
