package rabbitmq

import (
	"fmt"
	"time"

	"github.com/modulix-systems/goose-talk/contracts/rmqcontracts"
	"github.com/modulix-systems/goose-talk/logger"
	"github.com/rabbitmq/amqp091-go"
)

const (
	_defaultConnAttempts = 10
	_defaultConnTimeout  = time.Second
)

type RabbitMQ struct {
	connAttempts int
	connTimeout  time.Duration
	conn         *amqp091.Connection
	log          logger.Interface
}

func New(url string, log logger.Interface, options ...Option) (*RabbitMQ, error) {
	rmq := &RabbitMQ{connAttempts: _defaultConnAttempts, connTimeout: _defaultConnTimeout, log: log}

	for _, opt := range options {
		opt(rmq)
	}

	op := "rabbitmq.New"
	log = log.With("op", op)

	var err error
	for rmq.connAttempts > 0 {
		rmq.conn, err = amqp091.Dial(url)
		if err == nil {
			break
		}

		log.Error("connection failed, trying to reconnect", "attemptsLeft", rmq.connAttempts, "timeout", rmq.connTimeout, "err", err)

		time.Sleep(rmq.connTimeout)

		rmq.connAttempts--
	}

	if err != nil {
		return nil, fmt.Errorf("%s - no connection attempts left: %w", op, err)
	}

	return rmq, nil
}

func (rmq *RabbitMQ) NewChannel() (*amqp091.Channel, error) {
	return rmq.conn.Channel()
}

func (rmq *RabbitMQ) Close() error {
	return rmq.conn.Close()
}

func (rmq *RabbitMQ) QueueDeclare(queueContract rmqcontracts.Queue, channel *amqp091.Channel) (*amqp091.Queue, error) {
	op := "rabbitmq.RabbitMQ.QueueDeclare"
	log := rmq.log.With("op", op, "queue", queueContract.Name)

	queue, err := channel.QueueDeclare(queueContract.Name, queueContract.Durable, queueContract.Autodelete, queueContract.Exclusive, queueContract.NoWait, queueContract.Args)
	if err != nil {
		log.Error("failed to declare queue", "err", err)
		return nil, fmt.Errorf("%s - failed to declare queue: %w", op, err)
	}

	if queueContract.Binding != nil {
		if err := channel.QueueBind(
			queueContract.Name,
			queueContract.Binding.RoutingKey,
			queueContract.Binding.ExchangeNmae,
			queueContract.Binding.NoWait,
			queueContract.Binding.Args,
		); err != nil {
			return nil, err
		}
	}

	return &queue, nil
}

func (rmq *RabbitMQ) ExchangeDeclare(exchangeContract rmqcontracts.Exchange, channel *amqp091.Channel) error {
	return channel.ExchangeDeclare(
		exchangeContract.Name,
		exchangeContract.Kind,
		exchangeContract.Durable,
		exchangeContract.AutoDelete,
		exchangeContract.Internal,
		exchangeContract.NoWait,
		exchangeContract.Args,
	)
}
