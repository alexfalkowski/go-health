package checker

import "time"

func timeout(t time.Duration) time.Duration {
	if t == 0 {
		t = 30 * time.Second
	}

	return t
}
