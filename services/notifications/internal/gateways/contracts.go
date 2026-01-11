package gateways

import "context"

type (
	NotificationsRepo interface {
	}
	EmailSender interface {
		Send(ctx context.Context)
	}
)
