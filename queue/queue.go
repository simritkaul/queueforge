package queue

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/simritkaul/queueforge/internal"
)

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

	// Phase 1: Replay WAL
	if err := wal.Replay(q.applyWALRecord); err != nil {
		return nil, err;
	}

	// Phase 2: Requeue unfinished jobs
	for jobId, job :=  range q.jobs {
		if job.State == IN_PROGRESS {
			job.State = PENDING;
			q.pending = append(q.pending, jobId);
		}
	}

	return q, nil;
}

// Enqueues a new Job 
func (q *Queue) Enqueue (payload string) (string, error) {
	q.mu.Lock();
	defer q.mu.Unlock();

	jobId := generateJobId();

	data, err := internal.EncodeEnqueue(jobId, payload);

	if err != nil {
		return "", err;
	}

	// 1. Write to WAL
	if err := q.wal.Append(RecordEnqueue, data); err != nil {
		return "", err;
	}

	// 2. Update in-memory state
	job := &Job{
		ID: jobId,
		Payload: payload,
		State: PENDING,
	}

	q.jobs[jobId] = job;
	q.pending = append(q.pending, jobId);
	return jobId, nil;	
}

func generateJobId() string {
	return fmt.Sprintf("%d", time.Now().UnixNano());
}

// Dequeues the Job (FIFO)
func (q *Queue) Dequeue () (*Job, error) {
	q.mu.Lock();
	defer q.mu.Unlock();

	// If no pending jobs available
	if len(q.pending) == 0 {
		return nil, errors.New("no jobs available");
	}

	// Pop from the pending queue
	jobId := q.pending[0];
	q.pending = q.pending[1:];

	job, exists := q.jobs[jobId];
	if !exists {
		return nil, errors.New("no job found");
	}

	// Write dequeue to WAL
	if err := q.wal.Append(RecordDequeue, []byte(jobId)); err != nil {
		// Put the job back if write fails
		q.pending = append([]string{jobId}, q.pending...);
		return nil, err;
	}

	// Update in-memory state
	job.State = IN_PROGRESS;
	job.Attempts++;

	return job, nil;
}

// Marks a completed job as ACK (Acknowledged). Now that will never be processed.
func (q *Queue) Ack (jobId string) error {
	q.mu.Lock();
	defer q.mu.Unlock();

	job, exists := q.jobs[jobId];
	if !exists {
		return errors.New("no job found");
	}

	if job.State != IN_PROGRESS {
		return errors.New("job not in progress");
	}

	// Write ACK to WAL
	if err := q.wal.Append(RecordAck, []byte(jobId)); err != nil {
		return err;
	}

	// Update in-memory state
	job.State = ACKED;

	return nil;
}