package rabbit

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/gofrs/uuid"
	"github.com/rabbitmq/amqp091-go"
	"github.com/sunnyyssh/designing-software-cw3/order/internal/model"
)

type OrderService interface {
	SetOrderStatus(ctx context.Context, id uuid.UUID, status model.OrderStatus) (err error)
}

type Listener struct {
	service OrderService
	ch      *amqp091.Channel
	q       *amqp091.Queue
	logger  *slog.Logger
}

func NewListener(service OrderService, ch *amqp091.Channel, q *amqp091.Queue, logger *slog.Logger) *Listener {
	return &Listener{
		service: service,
		ch:      ch,
		q:       q,
		logger:  logger,
	}
}

func (l *Listener) Run(ctx context.Context) error {
	logger := l.logger.With("queue", l.q.Name)

	msgs, err := l.ch.Consume(
		l.q.Name, // queue
		"",       // consumer
		true,     // auto-ack (для простоты примера, в проде лучше false)
		false,    // exclusive
		false,    // no-local
		false,    // no-wait
		nil,      // args
	)
	if err != nil {
		return err
	}

LOOP:
	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				break LOOP
			}

			logger.Info("message received", "timestamp", msg.Timestamp)
			if err = l.handleMessage(ctx, msg, logger); err != nil {
				return err
			}

		case <-ctx.Done():
			break LOOP
		}
	}

	return nil
}

func (l *Listener) handleMessage(ctx context.Context, msg amqp091.Delivery, logger *slog.Logger) error {
	logger.DebugContext(ctx, "got message with body", "body", string(msg.Body))
	var orderMsg model.OrderServedMessage
	if err := json.Unmarshal(msg.Body, &orderMsg); err != nil {
		return err
	}

	return l.service.SetOrderStatus(ctx, orderMsg.ID, orderMsg.Status)
}
