package entity

import "time"

type PasskeyAuthTransport string

const (
	PASSKEY_AUTH_TRANSPORT_USB      PasskeyAuthTransport = "usb"
	PASSKEY_AUTH_TRANSPORT_NFC                           = "nfc"
	PASSKEY_AUTH_TRANSPORT_BLE                           = "ble"
	PASSKEY_AUTH_TRANSPORT_INTERNAL                      = "internal"
)

type TwoFADeliveryMethod int

const (
	TWO_FA_TELEGRAM TwoFADeliveryMethod = iota
	TWO_FA_EMAIL
	TWO_FA_SMS
	TWO_FA_TOTP_APP
)

func (m TwoFADeliveryMethod) String() string {
	switch m {
	case TWO_FA_TELEGRAM:
		return "telegram"
	case TWO_FA_EMAIL:
		return "email"
	case TWO_FA_SMS:
		return "sms"
	case TWO_FA_TOTP_APP:
		return "totp_app"
	default:
		return "unknown"
	}
}

var OtpDeliveryMethods = []TwoFADeliveryMethod{TWO_FA_EMAIL, TWO_FA_TELEGRAM, TWO_FA_TOTP_APP}

type (
	// OTP represents storage for verifications codes
	// Only one code can be present for one user/email.
	// In case if token already exists and requested for the same email again - it gets updated.
	// Therefore to check ttl expiration - UpdatedAt should be used
	// UserEmail and UserId are optional but at least one of them must be present
	OTP struct {
		Code      []byte
		UserEmail string
		UserId    int
		CreatedAt time.Time
		UpdatedAt time.Time
	}
	// TwoFactorAuth entity representing 2FA auth
	TwoFactorAuth struct {
		// user can have only one related 2fa entity
		UserId         int
		DeliveryMethod TwoFADeliveryMethod
		// could be whether user's telegram, email address, etc
		// or even optional (if required info is already present in user'entity) depending on DeliveryMethod.
		// The field can be optional e.g for email because it can be taken from user's acc
		Contact string
		// secret key required for otp generation if TOTP delivery method is used
		// stored as encrypted set of bytes
		TotpSecret []byte
		// indicates whether user has 2fa enabled. By default false
		Enabled bool
	}
	LoginToken struct {
		// ClientId is unique identifier for client which requested token
		ClientId         string
		Val              string
		ClientIdentity   *ClientIdentity
		ClientIdentityId int
		// AuthSessionId is optional and present only if token was approved
		// to retrieve auth session details
		AuthSessionId int
		AuthSession   *UserSession
		ExpiresAt     time.Time
	}

	PasskeyCredential struct {
		ID         []byte
		UserId     int
		PublicKey  []byte
		CreatedAt  time.Time
		LastUsedAt time.Time
		BackedUp   bool
		Transports []PasskeyAuthTransport
	}
)

func (otp *OTP) IsExpired(ttl time.Duration) bool {
	return time.Now().After(otp.UpdatedAt.Add(ttl))
}

func (l *LoginToken) IsExpired() bool {
	return time.Now().After(l.ExpiresAt)
}

func (l *LoginToken) IsApproved() bool {
	return l.AuthSessionId != 0
}
