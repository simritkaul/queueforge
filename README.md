# Queueforge

**Queueforge** is a minimal, durable message queue written in Go that focuses on
**correctness, reliability, and crash recovery** rather than throughput or scale.

It implements a **file-backed job queue** using a **Write-Ahead Log (WAL)** and
reconstructs in-memory state by **replaying the log on startup**, ensuring that
jobs are never lost even if the process crashes.

This project is inspired by the core design principles of production message
brokers such as Kafka, Amazon SQS, and RabbitMQ.

---

## Why Queueforge?

In real-world backend systems, asynchronous job processing must safely handle:

- Process crashes
- Worker failures
- Partial job execution
- Restarts without data loss

A simple in-memory queue fails under these conditions.  
Queueforge solves this by treating disk as the **source of truth** and memory as
**derived state**.

---

## Key Concepts Demonstrated

- Write-Ahead Logging (WAL)
- Durable state transitions
- At-least-once delivery semantics
- Crash recovery via log replay
- Explicit job lifecycle management
- Durability vs performance tradeoffs

---

## High-Level Architecture

```
Producer
    |
    v
Append-Only WAL ---> In-Memory Queue
    |
    v
Worker
    |
    v
ACK
```

- All state mutations are logged **before** being applied
- The WAL is append-only and ordered
- In-memory data structures can be rebuilt at any time

---

## Job Lifecycle

Each job progresses through a strict state machine:

### State Definitions

- **PENDING**

  - Job is enqueued and ready to be processed

- **IN_PROGRESS**

  - Job has been handed to a worker
  - Worker may crash or fail

- **ACKED**
  - Job has been successfully processed
  - Terminal state

### Allowed Transitions

PENDING → IN_PROGRESS

IN_PROGRESS → ACKED

IN_PROGRESS → PENDING (on crash / retry)

Disallowed transitions are intentionally prevented to maintain correctness.

---

## Write-Ahead Log (WAL)

Queueforge uses an **append-only WAL** to persist all job state transitions.

### WAL Record Types

```
ENQUEUE <jobID> <payload>
DEQUEUE <jobID>
ACK <jobID>
```

Each record represents a **single state transition** and is written to disk
before updating in-memory state.

---

## Crash Recovery

On startup, Queueforge performs recovery by:

1. Reading the WAL from the beginning
2. Replaying each record in order
3. Reconstructing job states in memory
4. Re-queuing any unfinished jobs

### Recovery Guarantees

- No jobs are lost
- Completed jobs are never reprocessed
- In-flight jobs are safely retried

---

## Delivery Semantics

Queueforge provides:

> **At-least-once delivery**

This means:

- Jobs may be processed more than once
- Jobs are never silently dropped

This tradeoff prioritizes correctness and durability over strict deduplication.

---

## Design Tradeoffs

- WAL replay is simpler and safer than snapshot-based recovery
- Disk persistence is prioritized over throughput
- Single-node design avoids unnecessary complexity

These tradeoffs are intentional and aligned with the learning goals of the
project.

---

## Project Structure

```
queueforge/
├── cmd/
│   └── main.go
├── queue/
│   ├── queue.go
│   ├── wal.go
│   ├── recovery.go
│   └── job.go
├── internal/
│   └── encoder.go
├── data/
│   └── queue.log
├── go.mod
└── README.md
```

---

## Limitations & Future Improvements

- Snapshotting and log compaction
- Exactly-once delivery using idempotency keys
- Multi-worker coordination
- Metrics and observability
- Networked producers and consumers

---

## Tech Stack

- Language: Go
- Persistence: Local file system
- Concurrency: Goroutines and channels
- Storage Model: Append-only Write-Ahead Log

---

## Key Takeaway

> **Correctness is a feature — and durability is a design choice.**

Queueforge is a focused exploration of how reliable systems handle failure
without sacrificing simplicity.
