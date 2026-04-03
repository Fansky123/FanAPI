package mq

import (
	"fanapi/internal/config"

	"github.com/nats-io/nats.go"
)

var Conn *nats.Conn

func Init(cfg *config.NATSConfig) error {
	var err error
	Conn, err = nats.Connect(cfg.URL)
	return err
}

func Publish(subject string, data []byte) error {
	return Conn.Publish(subject, data)
}

func Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
	return Conn.Subscribe(subject, handler)
}

func QueueSubscribe(subject, queue string, handler nats.MsgHandler) (*nats.Subscription, error) {
	return Conn.QueueSubscribe(subject, queue, handler)
}
