package main

import (
	"fmt"
	"log"
	"os"

	"github.com/simritkaul/queueforge/queue"
)

func main () {
	// Ensure data directory exists
	if err := os.MkdirAll("data", 0755); err != nil {
		log.Fatal(err);
	}

	// Open WAL
	wal, err := queue.OpenWAL("data/queue.log");
	if err != nil {
		log.Fatal(err);
	}
	defer wal.Close();

	// Create queue (along with Recovery)
	q, err := queue.NewQueue(wal);
	if err != nil {
		log.Fatal(err);
	}

	fmt.Println("QueueForge started");

	// Enqueue Jobs
	jobIds := []string{};
	for i := 1; i <= 3; i++ {
		jobId, err := q.Enqueue(fmt.Sprintf("task-%d", i));
		if err != nil {
			log.Fatal(err);
		}

		jobIds = append(jobIds, jobId);
		fmt.Println("Enqueued Job: ", jobId);
	}

	// Dequeue and Process Jobs
	for {
		job, err := q.Dequeue();
		if err != nil {
			fmt.Println("No more jobs to process");
			break;
		}

		fmt.Printf("Processing Job %s (payload = %s, attempt = %d)\n", job.ID, job.Payload, job.Attempts);

		// Simulate successful processing
		if err := q.Ack(job.ID); err != nil {
			log.Fatal(err);
		}

		fmt.Println("ACKed Job: ", job.ID);
	}

	fmt.Println("QueueForge shutting down")
}