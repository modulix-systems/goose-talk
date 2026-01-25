package rabbitmq

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/modulix-systems/goose-talk/contracts/rmqcontracts"
	"github.com/rabbitmq/amqp091-go"
)

type Handler interface {
	Handle(delivery amqp091.Delivery) error
}

type serverQueue struct {
	definition rmqcontracts.Queue
	handler    Handler
}

type Server struct {
	rmq          *RabbitMQ
	queues       []serverQueue
	channels     []*amqp091.Channel
	ServeErr     chan error
	wg           *sync.WaitGroup
	retryTimeout time.Duration
}

func NewServer(rmq *RabbitMQ, retryTimeout time.Duration) *Server {
	queues := make([]serverQueue, 0)
	serveErr := make(chan error)
	wg := new(sync.WaitGroup)
	channels := make([]*amqp091.Channel, 0)

	return &Server{rmq, queues, channels, serveErr, wg, retryTimeout}
}

func (s *Server) RegisterQueue(definition rmqcontracts.Queue, handler Handler) {
	s.queues = append(s.queues, serverQueue{definition, handler})
}

func (s *Server) serveQueue(serverQueue serverQueue) error {
	op := "rabbitmq.Server.serveQueue"
	log := s.rmq.log.With("op", op, "queue", serverQueue.definition.Name)

	channel, err := s.rmq.NewChannel()
	if err != nil {
		return fmt.Errorf("%s - create channel: %w", op, err)
	}
	s.channels = append(s.channels, channel)
	amqpQueue, err := s.rmq.QueueDeclare(serverQueue.definition, channel)
	if err != nil {
		return fmt.Errorf("%s - declare queue: %w", op, err)
	}
	deliveryChan, err := channel.Consume(amqpQueue.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("%s - consume queue: %w", op, err)
	}

	log.Info("start serving queue")

	for delivery := range deliveryChan {
		if err = serverQueue.handler.Handle(delivery); err != nil {
			delivery.Nack(false, true)
			log.Error("failed to handle delivery", "err", err, "retryTimeout", s.retryTimeout, "deliveryTag", delivery.DeliveryTag, "correlationId", delivery.CorrelationId)
			time.Sleep(s.retryTimeout)
		} else {
			delivery.Ack(false)
			log.Info("delivery acknowledged", "deliveryTag", delivery.DeliveryTag, "correlationId", delivery.CorrelationId)
		}
	}

	log.Info("Stop serving queue")

	return nil
}

func (s *Server) Run() {
	for _, queue := range s.queues {
		s.wg.Go(func() {
			if err := s.serveQueue(queue); err != nil {
				s.ServeErr <- err
				return
			}
		})
	}
}

func (s *Server) Stop(ctx context.Context) error {
	s.rmq.log.Info("Stopping rabbitmq server")

	for _, channel := range s.channels {
		channel.Close()
	}

	waitChannel := make(chan struct{})

	go func() {
		s.wg.Wait()
		close(waitChannel)
	}()

	select {
	case <-ctx.Done():
		s.rmq.log.Error("rabbitmq server graceful shutdown failed due to context timeout")
		return ctx.Err()
	case <-waitChannel:
		return nil
	}
}
