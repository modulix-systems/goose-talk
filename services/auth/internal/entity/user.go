package entity

import (
	"time"
)

type User struct {
	Id                 int    `json:"id"`
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
	PrivateKey         string `json:"-"`
}

func (u *User) Is2FAEnabled() bool {
	return u.TwoFactorAuth != nil && u.TwoFactorAuth.Enabled
}
