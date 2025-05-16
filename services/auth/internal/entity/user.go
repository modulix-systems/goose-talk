package entity

import "time"

type TwoFADeliveryMethod = int

const (
	TELEGRAM TwoFADeliveryMethod = iota
	EMAIL
	SMS
)

type (
	// SignUpCode represents storage for email verifications codes
	// Only one code can be present for one email
	SignUpCode struct {
		Code      string
		Email     string
		CreatedAt time.Time
	}
	User struct {
		ID         int
		Username   string
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
	// TwoFactorAuth entity representing 2FA auth (if user has enabled it)
	TwoFactorAuth struct {
		UserId         int
		DeliveryMethod TwoFADeliveryMethod
		// could be whether user's telegram, email address or phone number
		// depending on DeliveryMethod.
		// The field can be optional e.g for email because it can be taken from user's acc
		contact   string
		otpSecret string
		Enabled   bool
	}
)
