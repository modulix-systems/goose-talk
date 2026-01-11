package entity

type TwoFATransport string

const (
	TWO_FA_EMAIL    TwoFATransport = "email"
	TWO_FA_SMS      TwoFATransport = "sms"
	TWO_FA_TELEGRAM TwoFATransport = "telegram"
	TWO_FA_TOTP_APP TwoFATransport = "totp_app"
)

var OtpTransports = []TwoFATransport{TWO_FA_EMAIL, TWO_FA_TELEGRAM, TWO_FA_TOTP_APP}

type (
	// OTP represents storage for verifications codes
	// Only one code can be present for one user/email.
	// UserEmail and UserId are optional but at least one of them must be present
	OTP struct {
		Code      []byte
		UserEmail string `json:"user_email"`
		UserId    int    `json:"user_id"`
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
)
