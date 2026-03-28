package rabbit

import (
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	ch  *amqp.Channel
	log *slog.Logger
}

func NewConsumer(ch *amqp.Channel, log *slog.Logger) *Consumer {
	return &Consumer{
		ch:  ch,
		log: log,
	}
}

func (c *Consumer) Consume(queue string, handler func(msg []byte) error) error {
	msgs, err := c.ch.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		for msg := range msgs {
			if err := handler(msg.Body); err != nil {
				continue
			}
			if err := msg.Ack(false); err != nil {
				c.log.Error("ack error")
			}

		}
	}()
	return nil
}
