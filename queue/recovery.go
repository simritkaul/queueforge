package queue

import (
	"errors"

	"github.com/simritkaul/queueforge/internal"
)

// Handler to update jobs and pending in Recovery
func (q *Queue) applyWALRecord(record WALRecord) error {
	q.mu.Lock();
	defer q.mu.Unlock();

	switch record.Type {
	case RecordEnqueue:
			jobId, payload, err := internal.DecodeEnqueue(record.Data);
			if err != nil {
				return err;
			}

			// Create job if it doesn't exist
			if _, exists := q.jobs[jobId]; !exists {
				q.jobs[jobId] = &Job{
					ID: jobId,
					Payload: payload,
					State: PENDING,
				}
				q.pending = append(q.pending, jobId);
			}

		case RecordDequeue:
			jobId := string(record.Data);
			job, exists := q.jobs[jobId];
			if !exists {
				return errors.New("DEQUEUE for unknown job");
			}

			job.State = IN_PROGRESS;
			job.Attempts++;

		case RecordAck:
			jobId := string(record.Data);
			job, exists := q.jobs[jobId];
			if !exists {
				return errors.New("ACK for unknown job");
			}
			
			// soft delete
			job.State = ACKED;

		default:
			return errors.New("Unknown WAL record type");
	}

	return nil;
}