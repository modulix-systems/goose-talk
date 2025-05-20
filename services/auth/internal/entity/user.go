package entity

import (
	"time"
)

type (
	User struct {
		ID            int
		Username      string
		Password      []byte `json:"-"`
		Email         string
		FirstName     string
		LastName      string
		PhotoUrl      string
		Friends       []User
		CreatedAt     time.Time `db:"created_at" json:"created_at"`
		UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
		LastSeenAt    time.Time
		IsActive      bool
		BirthDate     time.Time
		AboutMe       string
		TwoFactorAuth *TwoFactorAuth
	}
	// UserSession stores information about user's session within single device
	// Allows to forbid access to user if his ip is not in a list of user's active sessions
	UserSession struct {
		ID       int
		UserId   int
		Location string
		// IP serves role of unique session identifier
		IP         string
		LastSeenAt time.Time
		CreatedAt  time.Time
		DeviceInfo string
		// nil by default if session is active
		DeactivatedAt time.Time
		AccessToken   string // unique
	}
)

func (s *UserSession) IsActive() bool {
	return s.DeactivatedAt.IsZero()
}

func (u *User) Is2FAEnabled() bool {
	return u.TwoFactorAuth != nil && u.TwoFactorAuth.Enabled
}
