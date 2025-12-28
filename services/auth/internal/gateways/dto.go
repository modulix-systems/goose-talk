package gateways

import "time"

type (
	TelegramMsg struct {
		DateSent time.Time
		Text     string
		ChatId   string
	}
)
