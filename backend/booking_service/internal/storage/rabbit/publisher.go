package rabbit

import (
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PublisherInterface interface {
	Publish(queue string, msg interface{}) error
}

type Publisher struct {
	ch *amqp.Channel
}

func NewPublisher(ch *amqp.Channel) *Publisher {
	return &Publisher{ch: ch}
}

func (p *Publisher) Publish(queue string, msg interface{}) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return p.ch.Publish(
		"",
		queue,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}
