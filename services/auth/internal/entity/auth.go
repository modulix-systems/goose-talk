package entity

import "time"

type (
	// AuthSession is a rolling auth session
	// which stores information about user's session within single device
	// Allows to forbid access to user if his ip is not in a list of user's active sessions
	AuthSession struct {
		ID          string    `json:"id"`
		UserId      int       `json:"user_id"`
		LastSeenAt  time.Time `json:"last_seen_at"`
		CreatedAt   time.Time `json:"created_at"`
		IsLongLived bool

		// Login metadata
		Location   string `json:"location"`
		IPAddr     string `json:"ip_addr"`
		DeviceInfo string `json:"device_info"`
	}

	// Entity for QR code login flow. Inspired from telegram (https://core.telegram.org/api/qr-login)
	QRCodeLoginToken struct {
		Value string `json:"value"`
		// ClientId is unique identifier for client which requested token
		ClientId string

		// Login metadata
		IPAddr     string
		DeviceInfo string
	}
)
