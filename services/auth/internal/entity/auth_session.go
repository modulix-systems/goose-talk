package entity

import "time"

// AuthSession is a rolling auth session
// which stores information about user's login within single device
type AuthSession struct {
	Id          string    `json:"id"`
	UserId      int       `json:"user_id"`
	LastSeenAt  time.Time `json:"last_seen_at"`
	CreatedAt   time.Time `json:"created_at"`
	IsLongLived bool

	// Login metadata
	Location   string `json:"location"`
	IpAddr     string `json:"ip_addr"`
	DeviceInfo string `json:"device_info"`
}
