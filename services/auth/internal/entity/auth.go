package entity

import "time"

type (
	// UserSession is a rolling auth session
	// which stores information about user's session within single device
	// Allows to forbid access to user if his ip is not in a list of user's active sessions
	UserSession struct {
		ID         string    `json:"id"`
		User       *User     `json:"user"`
		UserId     int       `json:"user_id"`
		LastSeenAt time.Time `json:"last_seen_at"`
		// ExpiresAt should be automatically updated when user interacts within session
		ExpiresAt        time.Time       `json:"expires_at"`
		CreatedAt        time.Time       `json:"created_at"`
		// nil by default if session is active
		DeactivatedAt time.Time `json:"deactivated_at"`

		// Login metadata
		Location   string `json:"location"`
		IPAddr     string `json:"ip_addr"`
		DeviceInfo string `json:"device_info"`
	}

	// Entity for QR code login flow. Inspired from telegram (https://core.telegram.org/api/qr-login)
	QRCodeLoginToken struct {
		// ClientId is unique identifier for client which requested token
		ClientId string `json:"client_id"`
		Value    string `json:"val"`
		// AuthSessionId is optional and present only if token was approved
		// to retrieve auth session details
		AuthSessionId int          `json:"auth_session_id"`
		AuthSession   *UserSession `json:"auth_session"`

		// Login metadata
		Location   string `json:"location"`
		IPAddr     string `json:"ip_addr"`
		DeviceInfo string `json:"device_info"`
	}
)
