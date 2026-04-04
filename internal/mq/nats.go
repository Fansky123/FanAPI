package mq

import (
	"fmt"
	"log"
	"strings"
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

	// Clean up any stale consumers whose filter subjects no longer match
	// (e.g. old "script-workers" bound to "task.image.*" after migrating to "task.>").
	for info := range JS.Consumers(streamName) {
		name := info.Name
		filter := info.Config.FilterSubject
		// A consumer is stale if it claims to be a workers-* consumer but its filter
		// no longer matches any of the expected patterns.
		if strings.HasPrefix(name, "workers-") && filter != "" && filter != streamSubj {
			expected := "workers-" + strings.NewReplacer(
				"task.", "", ".*", "", ".", "-", ">", "all", "*", "any",
			).Replace(filter)
			if name != expected {
				_ = JS.DeleteConsumer(streamName, name)
				log.Printf("[mq] deleted stale consumer %q (filter: %q)", name, filter)
			}
		}
	}
	return nil
}

// Publish persists a message durably to JetStream before returning.
// The call blocks until the NATS server acknowledges the write to the stream.
func Publish(subject string, data []byte) error {
	_, err := JS.Publish(subject, data)
	return err
}

// QueueSubscribe creates a durable JetStream push consumer shared across a queue group.
// Each message is delivered to exactly one subscriber. If the subscriber crashes without
// calling msg.Ack(), the message is automatically redelivered after workerAckWait.
func QueueSubscribe(subject, queue string, handler nats.MsgHandler) (*nats.Subscription, error) {
	return JS.QueueSubscribe(
		subject, queue, handler,
		nats.Durable(queue),
		nats.AckExplicit(),
		nats.AckWait(workerAckWait),
		nats.MaxDeliver(workerMaxDeliver),
	)
}

// Subscribe creates a core NATS subscription (non-persistent, for ancillary use cases).
func Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
	return Conn.Subscribe(subject, handler)
}
