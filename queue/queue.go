package queue

import "sync"

type Queue struct {
	wal *WAL;
	jobs	map[string]*Job;
	pending []string;
	mu sync.Mutex;
}

// Initialises a new Queue
func NewQueue (wal *WAL) (*Queue, error) {
	q := &Queue{
		wal: wal,
		jobs: make(map[string]*Job),
		pending: make([]string, 0),
	}

	// Recovery will happen here
	return q, nil;
}