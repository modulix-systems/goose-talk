package entity

import (
	"time"
)

type (
	User struct {
		ID                 int
		Username           string
		Password           []byte `json:"-"`
		Email              string
		FirstName          string
		LastName           string
		PhotoUrl           string
		Friends            []User
		CreatedAt          time.Time `json:"created_at" db:"created_at"`
		UpdatedAt          time.Time `json:"updated_at" db:"updated_at"`
		LastSeenAt         time.Time
		IsActive           bool
		BirthDate          time.Time
		AboutMe            string
		TwoFactorAuth      *TwoFactorAuth
		PasskeyCredentials []PasskeyCredential
	}
	// UserSession stores information about user's session within single device
	// Allows to forbid access to user if his ip is not in a list of user's active sessions
	UserSession struct {
		ID               int
		UserId           int
		LastSeenAt       time.Time
		CreatedAt        time.Time
		ClientIdentity   *ClientIdentity
		ClientIdentityId int
		// nil by default if session is active
		DeactivatedAt time.Time
		AccessToken   string // unique
	}
	ClientIdentity struct {
		ID         int
		Location   string
		IPAddr     string
		DeviceInfo string
	}
)

func (s *UserSession) IsActive() bool {
	return s.DeactivatedAt.IsZero()
}

func (u *User) Is2FAEnabled() bool {
	return u.TwoFactorAuth != nil && u.TwoFactorAuth.Enabled
}
