package rabbit

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	ch *amqp091.Channel
	q  *amqp091.Queue
}

func NewPublisher(ch *amqp091.Channel, q *amqp091.Queue) *Publisher {
	return &Publisher{
		ch: ch,
		q:  q,
	}
}

func (p *Publisher) Publish(ctx context.Context, msgs ...any) (err error) {
	for _, msg := range msgs {
		body, err := json.Marshal(msg)
		if err != nil {
			return err
		}

		err = p.ch.PublishWithContext(ctx,
			"",       // exchange (используем default)
			p.q.Name, // routing key (имя очереди)
			false,    // mandatory
			false,    // immediate
			amqp091.Publishing{
				Timestamp:   time.Now(),
				ContentType: "application/json",
				Body:        []byte(body),
			},
		)
		if err != nil {
			return err
		}
	}

	return err
}
