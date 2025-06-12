package entity

import "time"

type PasskeyAuthTransport string

const (
	PASSKEY_AUTH_TRANSPORT_USB      PasskeyAuthTransport = "usb"
	PASSKEY_AUTH_TRANSPORT_NFC      PasskeyAuthTransport = "nfc"
	PASSKEY_AUTH_TRANSPORT_BLE      PasskeyAuthTransport = "ble"
	PASSKEY_AUTH_TRANSPORT_INTERNAL PasskeyAuthTransport = "internal"
)

type TwoFATransport string

const (
	TWO_FA_TELEGRAM TwoFATransport = "telegram"
	TWO_FA_EMAIL    TwoFATransport = "email"
	TWO_FA_SMS      TwoFATransport = "sms"
	TWO_FA_TOTP_APP TwoFATransport = "totp_app"
)

var OtpTransports = []TwoFATransport{TWO_FA_EMAIL, TWO_FA_TELEGRAM, TWO_FA_TOTP_APP}

type (
	// OTP represents storage for verifications codes
	// Only one code can be present for one user/email.
	// In case if token already exists and requested for the same email again - it gets updated.
	// Therefore to check ttl expiration - UpdatedAt should be used
	// UserEmail and UserId are optional but at least one of them must be present
	OTP struct {
		Code      []byte
		UserEmail string    `json:"user_email"`
		UserId    int       `json:"user_id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
	// TwoFactorAuth entity representing 2FA auth
	TwoFactorAuth struct {
		// user can have only one related 2fa entity
		UserId    int            `json:"user_id"`
		Transport TwoFATransport `json:"transport"`
		// could be whether user's telegram, email address, etc
		// or even optional (if required info is already present in user'entity) depending on Transport.
		// The field can be optional e.g for email because it can be taken from user's acc
		Contact string `json:"contact"`
		// secret key required for otp generation if TOTP delivery method is used
		// stored as encrypted set of bytes
		TotpSecret []byte `json:"totp_secret"`
		// indicates whether user has 2fa enabled.
		// By default true, but can be disabled on user's demand
		Enabled bool `json:"enabled"`
	}

	LoginToken struct {
		// ClientId is unique identifier for client which requested token
		ClientId         string          `json:"client_id"`
		Val              string          `json:"val"`
		ClientIdentity   *ClientIdentity `json:"client_identity"`
		ClientIdentityId int             `json:"client_identity_id"`
		// AuthSessionId is optional and present only if token was approved
		// to retrieve auth session details
		AuthSessionId int          `json:"auth_session_id"`
		AuthSession   *UserSession `json:"auth_session"`
		ExpiresAt     time.Time    `json:"expires_at"`
	}

	PasskeyCredential struct {
		ID         string                 `json:"id"`
		UserId     int                    `json:"user_id"`
		PublicKey  []byte                 `json:"public_key"`
		CreatedAt  time.Time              `json:"created_at"`
		LastUsedAt time.Time              `json:"last_used_at"`
		BackedUp   bool                   `json:"backed_up"`
		Transports []PasskeyAuthTransport `json:"transports"`
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
