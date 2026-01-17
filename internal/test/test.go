package test

import (
	"cmp"
	"fmt"
	"os"
)

// StatusURL for test.
func StatusURL(status string) string {
	port := cmp.Or(os.Getenv("STATUS_PORT"), "6000")
	return fmt.Sprintf("http://localhost:%s/v1/status/%s", port, status)
}
