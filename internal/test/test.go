// Package test contains internal test helpers.
package test

import (
	"cmp"
	"fmt"
	"os"
)

// StatusURL returns the URL for the local status test service.
func StatusURL(status string) string {
	port := cmp.Or(os.Getenv("STATUS_PORT"), "6000")
	return fmt.Sprintf("http://localhost:%s/v1/status/%s", port, status)
}
