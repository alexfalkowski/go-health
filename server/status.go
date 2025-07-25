package server

const (
	// Started status.
	Started = Status("started")

	// Stopped status.
	Stopped = Status("stopped")
)

// Status of the server.
type Status string

// IsEmpty returns true if a status is set.
func (s Status) IsEmpty() bool {
	return len(s) == 0
}

// IsStarted return true if it is started.
func (s Status) IsStarted() bool {
	return s == Started
}

// IsStopped return true if it is stopped.
func (s Status) IsStopped() bool {
	return s == Stopped
}
