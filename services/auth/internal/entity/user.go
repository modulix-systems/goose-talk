package entity

import (
	"time"
)

type User struct {
	Id                 int    `json:"id"`
	Username           string `json:"username"`
	Password           []byte `json:"-"`
	Email              string `json:"email"`
	PhoneNumber        string `json:"phone_number"`
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
	PrivateKey         string `json:"-"`
	Language           string `json:"language"`
}

func (u *User) Is2FAEnabled() bool {
	return u.TwoFactorAuth != nil && u.TwoFactorAuth.Enabled
}

func (u *User) GetDisplayName() string {
	displayName := u.Username
	if u.FirstName != "" {
		displayName = u.FirstName
		if u.LastName != "" {
			displayName = displayName + " " + u.LastName
		}
	}

	return displayName
}
