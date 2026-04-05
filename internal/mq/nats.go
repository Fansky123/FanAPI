package mq

import (
	"fmt"
	"log"
	"time"

	"fanapi/internal/config"

	"github.com/nats-io/nats.go"
)

const (
	streamName = "TASKS"
	streamSubj = "task.>"

	// workerAckWait: worker must ACK within this window or the message is redelivered.
	// Set to well above the longest expected task processing time.
	workerAckWait = 10 * time.Minute

	// workerMaxDeliver: max delivery attempts before the message is dropped (not retried forever).
	workerMaxDeliver = 3
)

var (
	Conn *nats.Conn
	JS   nats.JetStreamContext
)

func Init(cfg *config.NATSConfig) error {
	var err error
	Conn, err = nats.Connect(cfg.URL)
	if err != nil {
		return fmt.Errorf("nats connect: %w", err)
	}
	JS, err = Conn.JetStream()
	if err != nil {
		return fmt.Errorf("jetstream context: %w", err)
	}
	return nil
}

// EnsureStream creates or updates the persistent TASKS stream.
// Must be called once on startup in every process that publishes or subscribes to task subjects.
func EnsureStream() error {
	cfg := &nats.StreamConfig{
		Name:     streamName,
		Subjects: []string{streamSubj},
		// WorkQueuePolicy: each message is delivered to exactly one consumer and deleted on ACK.
		Retention: nats.WorkQueuePolicy,
		// FileStorage: messages survive NATS server restarts.
		Storage: nats.FileStorage,
		// Drop messages older than 24h to prevent unbounded backlog on extended outages.
		MaxAge:   24 * time.Hour,
		Replicas: 1,
	}

	_, err := JS.StreamInfo(streamName)
	if err == nats.ErrStreamNotFound {
		if _, err = JS.AddStream(cfg); err != nil {
			return fmt.Errorf("create TASKS stream: %w", err)
		}
		log.Printf("[mq] JetStream stream %q created (FileStorage, WorkQueuePolicy)", streamName)
	} else if err != nil {
		return fmt.Errorf("stream info: %w", err)
	} else {
		if _, err = JS.UpdateStream(cfg); err != nil {
			return fmt.Errorf("update TASKS stream: %w", err)
		}
		log.Printf("[mq] JetStream stream %q confirmed", streamName)
	}

	// Note: consumer cleanup is intentionally NOT done here.
	// Only the worker process should purge stale consumers (via mq.PurgeConsumers)
	// before subscribing. Running cleanup here would delete the worker's active
	// consumer whenever the server restarts.
	return nil
}

// Publish persists a message durably to JetStream before returning.
// The call blocks until the NATS server acknowledges the write to the stream.
func Publish(subject string, data []byte) error {
	_, err := JS.Publish(subject, data)
	return err
}

// QueueSubscribe creates a durable JetStream pull consumer and starts a background goroutine
// to continuously fetch messages. Pull consumers work correctly on WorkQueue streams and
// survive worker restarts without the "filtered consumer not unique" error that push
// consumers suffer on WorkQueue streams.
//
// Each fetched message is dispatched in its own goroutine (safe: handler is idempotent
// and operates on independent task IDs). The goroutine exits when the subscription is
// closed (sub.Unsubscribe / server shutdown).
func QueueSubscribe(subject, queue string, handler nats.MsgHandler) (*nats.Subscription, error) {
	sub, err := JS.PullSubscribe(
		subject,
		queue,
		nats.AckExplicit(),
		nats.AckWait(workerAckWait),
		nats.MaxDeliver(workerMaxDeliver),
	)
	if err != nil {
		return nil, fmt.Errorf("pull subscribe %s/%s: %w", subject, queue, err)
	}

	go func() {
		for {
			msgs, fetchErr := sub.Fetch(10, nats.MaxWait(5*time.Second))
			if fetchErr != nil {
				if !sub.IsValid() {
					return // subscription closed, exit cleanly
				}
				// ErrTimeout is normal when the queue is empty
				if fetchErr != nats.ErrTimeout {
					log.Printf("[mq] fetch error (%s): %v", subject, fetchErr)
					time.Sleep(time.Second)
				}
				continue
			}
			for _, msg := range msgs {
				go handler(msg)
			}
		}
	}()

	return sub, nil
}

// Subscribe creates a core NATS subscription (non-persistent, for ancillary use cases).
func Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
	return Conn.Subscribe(subject, handler)
}

// PurgeConsumers removes all consumers from the TASKS stream.
// Must be called once from the worker process before QueueSubscribe to clear stale
// consumers left over from previous runs (prevents "filtered consumer not unique"
// on WorkQueue streams). Must NOT be called from the server process.
func PurgeConsumers() {
	for info := range JS.Consumers(streamName) {
		if delErr := JS.DeleteConsumer(streamName, info.Name); delErr == nil {
			log.Printf("[mq] purged stale consumer %q from stream %s", info.Name, streamName)
		}
	}
}
