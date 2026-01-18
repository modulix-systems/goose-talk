package rabbitmq

import (
	"context"
	"fmt"
	"sync"

	"github.com/modulix-systems/goose-talk/contracts/rmqcontracts"
	"github.com/rabbitmq/amqp091-go"
)

type Handler interface {
	Handle(delivery amqp091.Delivery)
}

type serveableQueue struct {
	definition rmqcontracts.Queue
	handler    Handler
}

type Server struct {
	rmq      *RabbitMQ
	queues   []serveableQueue
	channels []*amqp091.Channel
	ServeErr chan error
	wg       *sync.WaitGroup
}

func NewServer(rmq *RabbitMQ) *Server {
	queues := make([]serveableQueue, 0)
	serveErr := make(chan error)
	wg := new(sync.WaitGroup)
	channels := make([]*amqp091.Channel, 0)

	return &Server{rmq, queues, channels, serveErr, wg}
}

func (s *Server) RegisterQueue(definition rmqcontracts.Queue, handler Handler) {
	s.queues = append(s.queues, serveableQueue{definition, handler})
}

func (s *Server) serveQueue(serveableQueue serveableQueue) error {
	channel, err := s.rmq.NewChannel()
	if err != nil {
		return fmt.Errorf("rabbitmq - RabbitMQ.ServiceQueue - create channel: %w", err)
	}
	s.channels = append(s.channels, channel)
	queue, err := s.rmq.QueueDeclare(serveableQueue.definition, channel)
	if err != nil {
		return fmt.Errorf("rabbitmq - RabbitMQ.ServiceQueue - declare queue: %w", err)
	}
	deliveryChan, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("rabbitmq - RabbitMQ.ServiceQueue - consume queue: %w", err)
	}

	for delivery := range deliveryChan {
		serveableQueue.handler.Handle(delivery)
		delivery.Ack(false)
	}

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
		return ctx.Err()
	case <-waitChannel:
		return nil
	}
}
