package queue

type JobState int

const (
	PENDING JobState = iota
	IN_PROGRESS
	ACKED
)

type Job struct {
	ID       string
	Payload  string
	State    JobState
	Attempts int
}