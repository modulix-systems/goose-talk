module github.com/modulix-systems/goose-talk/rabbitmq

replace github.com/modulix-systems/goose-talk/contracts => ../contracts

go 1.25.5

require (
	github.com/modulix-systems/goose-talk/contracts v0.0.0-00010101000000-000000000000
	github.com/rabbitmq/amqp091-go v1.10.0
)
