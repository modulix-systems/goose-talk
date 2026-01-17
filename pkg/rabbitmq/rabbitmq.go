package rabbitmq

import (
	"fmt"
	"log"
	"time"

	"github.com/modulix-systems/goose-talk/contracts/rmqcontracts"
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
}

func New(url string, options ...Option) (*RabbitMQ, error) {
	rmq := &RabbitMQ{connAttempts: _defaultConnAttempts, connTimeout: _defaultConnTimeout}

	for _, opt := range options {
		opt(rmq)
	}

	var err error
	for rmq.connAttempts > 0 {
		rmq.conn, err = amqp091.Dial(url)
		if err == nil {
			break
		}

		log.Printf("RabbitMQ is trying to connect, attempts left: %d", rmq.connAttempts)

		time.Sleep(rmq.connTimeout)

		rmq.connAttempts--
	}

	if err != nil {
		return nil, fmt.Errorf("rabbitmq - New - connAttempts == 0: %w", err)
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
	queue, err := channel.QueueDeclare(queueContract.Name, queueContract.Durable, queueContract.Autodelete, queueContract.Exclusive, queueContract.NoWait, queueContract.Args)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq - RabbitMQ.QueueDeclare - channel.QueueDeclare: %w", err)
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