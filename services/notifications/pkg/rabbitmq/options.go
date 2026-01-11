package rabbitmq

import "time"

type Option func(*RabbitMQ)

// ConnAttempts -.
func ConnAttempts(attempts int) Option {
	return func(c *RabbitMQ) {
		c.connAttempts = attempts
	}
}

// ConnTimeout -.
func ConnTimeout(timeout time.Duration) Option {
	return func(c *RabbitMQ) {
		c.connTimeout = timeout
	}
}
