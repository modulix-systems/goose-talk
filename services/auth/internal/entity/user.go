package entity

import (
	"time"
)

type (
	User struct {
		ID                 int    `json:"id"`
		Username           string `json:"username"`
		Password           []byte `json:"-"`
		Email              string `json:"email"`
		FirstName          string `json:"first_name"`
		LastName           string `json:"last_name"`
		PhotoUrl           string `json:"photo_url"`
		Friends            []User
		CreatedAt          time.Time      `json:"created_at" db:"created_at"`
		UpdatedAt          time.Time      `json:"updated_at" db:"updated_at"`
		IsActive           bool           `json:"is_active"`
		BirthDate          time.Time      `json:"birth_date"`
		AboutMe            string         `json:"about_me"`
		TwoFactorAuth      *TwoFactorAuth `json:"two_factor_auth" db:"-"`
		PasskeyCredentials []PasskeyCredential
	}
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
		ClientIdentity   *ClientIdentity `json:"client_identity"`
		ClientIdentityId int             `json:"client_identity_id"`
		// nil by default if session is active
		DeactivatedAt time.Time `json:"deactivated_at"`
	}

	ClientIdentity struct {
		ID         int    `json:"id"`
		Location   string `json:"location"`
		IPAddr     string `json:"ip_addr"`
		DeviceInfo string `json:"device_info"`
	}
)

func (s *UserSession) IsActive() bool {
	return s.DeactivatedAt.IsZero()
}

func (u *User) Is2FAEnabled() bool {
	return u.TwoFactorAuth != nil && u.TwoFactorAuth.Enabled
}
