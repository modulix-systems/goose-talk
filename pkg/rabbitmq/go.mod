module github.com/modulix-systems/goose-talk/rabbitmq

replace github.com/modulix-systems/goose-talk/contracts => ../contracts

go 1.25.5

require (
	github.com/modulix-systems/goose-talk/contracts v0.0.0-00010101000000-000000000000
	github.com/modulix-systems/goose-talk/logger v0.0.0-00010101000000-000000000000
	github.com/rabbitmq/amqp091-go v1.10.0
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.19 // indirect
	github.com/rs/zerolog v1.34.0 // indirect
	golang.org/x/sys v0.12.0 // indirect
)

replace github.com/modulix-systems/goose-talk/logger => ../../pkg/logger
