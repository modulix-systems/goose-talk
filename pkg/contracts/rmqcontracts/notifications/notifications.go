package notifications

import (
	"github.com/modulix-systems/goose-talk/contracts/rmqcontracts"
)

type Contracts struct {
	Queues Queues
}

type Queues struct {
	Emails        rmqcontracts.Queue
	Notifications rmqcontracts.Queue
}

func New() *Contracts {
	return &Contracts{
		Queues: Queues{
			Emails: rmqcontracts.Queue{
				Name:    "emails",
				Durable: true,
			},
			Notifications: rmqcontracts.Queue{
				Name:    "emails",
				Durable: true,
			},
		},
	}

}
