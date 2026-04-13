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

	resultStreamName = "RESULTS"
	resultStreamSubj = "result.>"

	// workerAckWait: worker must ACK within this window or the message is redelivered.
	// Set to well above the longest expected task processing time.
	workerAckWait = 10 * time.Minute

	// workerMaxDeliver: max delivery attempts before the message is dropped (not retried forever).
	workerMaxDeliver = 3
)

var (
	Conn  *nats.Conn
	JS    nats.JetStreamContext
	mqCfg *config.NATSConfig
)

func Init(cfg *config.NATSConfig) error {
	mqCfg = cfg
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

// EnsureStream creates or updates the persistent TASKS and RESULTS JetStream streams.
// Must be called once on startup in every process that uses NATS.
func EnsureStream() error {
	storage := nats.FileStorage
	replicas := 1
	if mqCfg != nil {
		if mqCfg.MemoryStorage {
			storage = nats.MemoryStorage
		}
		if mqCfg.Replicas > 1 {
			replicas = mqCfg.Replicas
		}
	}
	if err := ensureOneStream(&nats.StreamConfig{
		Name:      streamName,
		Subjects:  []string{streamSubj},
		Retention: nats.WorkQueuePolicy,
		Storage:   storage,
		MaxAge:    24 * time.Hour,
		Replicas:  replicas,
	}); err != nil {
		return err
	}
	return ensureOneStream(&nats.StreamConfig{
		Name:      resultStreamName,
		Subjects:  []string{resultStreamSubj},
		Retention: nats.WorkQueuePolicy,
		Storage:   storage,
		MaxAge:    24 * time.Hour,
		Replicas:  replicas,
	})
}

func ensureOneStream(cfg *nats.StreamConfig) error {
	_, err := JS.StreamInfo(cfg.Name)
	if err == nats.ErrStreamNotFound {
		if _, err = JS.AddStream(cfg); err != nil {
			return fmt.Errorf("create %s stream: %w", cfg.Name, err)
		}
		log.Printf("[mq] JetStream stream %q created", cfg.Name)
	} else if err != nil {
		return fmt.Errorf("stream info (%s): %w", cfg.Name, err)
	} else {
		if _, err = JS.UpdateStream(cfg); err != nil {
			return fmt.Errorf("update %s stream: %w", cfg.Name, err)
		}
		log.Printf("[mq] JetStream stream %q confirmed", cfg.Name)
	}
	return nil
}

// PublishResult durably publishes a worker result to the RESULTS stream.
func PublishResult(subject string, data []byte) error {
	_, err := JS.Publish(subject, data)
	return err
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
// maxConcurrent limits the number of goroutines running handler concurrently.
// When the limit is reached, the fetch loop blocks — providing true backpressure:
// no new messages are pulled from the queue until a slot is freed.
// Pass 0 for unlimited concurrency.
func QueueSubscribe(subject, queue string, handler nats.MsgHandler, maxConcurrent int) (*nats.Subscription, error) {
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

	var sem chan struct{}
	if maxConcurrent > 0 {
		sem = make(chan struct{}, maxConcurrent)
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
				if sem != nil {
					// 先占槽再 spawn goroutine：fetch 循环在此阻塞，
					// 达到上限时不再从队列拉取新消息，形成真正的背压。
					sem <- struct{}{}
				}
				go func(m *nats.Msg) {
					if sem != nil {
						defer func() { <-sem }()
					}
					handler(m)
				}(msg)
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
