package entity

import (
	"time"
)

type TwoFADeliveryMethod = int

const (
	TELEGRAM TwoFADeliveryMethod = iota
	EMAIL
	SMS
	TOTP_APP
)

type (
	// OTP represents storage for email verifications codes
	// Only one code can be present for one email.
	// In case if token already exists and requested for the same email again - it gets updated.
	// So UpdatedAt should be used to check ttl expiration
	OTP struct {
		Code      []byte
		UserEmail string
		CreatedAt time.Time
		UpdatedAt time.Time
	}
	User struct {
		ID         int
		Username   string
		Password   []byte `json:"-"`
		Email      string
		FirstName  string
		LastName   string
		PhotoUrl   string
		Friends    []User
		CreatedAt  time.Time `db:"created_at" json:"created_at"`
		UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
		LastSeenAt time.Time
		IsActive   bool
		BirthDate  time.Time
		AboutMe    string
	}
	// TwoFactorAuth entity representing 2FA auth
	TwoFactorAuth struct {
		UserId         int
		DeliveryMethod TwoFADeliveryMethod
		// could be whether user's telegram, email address or phone number
		// depending on DeliveryMethod.
		// The field can be optional e.g for email because it can be taken from user's acc
		contact string
		// secret key required for otp generation
		otpSecret string
		// indicates whether user has 2fa enabled. By default false
		Enabled bool
	}
)
