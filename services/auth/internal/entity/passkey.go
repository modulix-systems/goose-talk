package entity

import "time"

type PasskeyAuthTransport string

const (
	PASSKEY_AUTH_TRANSPORT_USB      PasskeyAuthTransport = "usb"
	PASSKEY_AUTH_TRANSPORT_NFC      PasskeyAuthTransport = "nfc"
	PASSKEY_AUTH_TRANSPORT_BLE      PasskeyAuthTransport = "ble"
	PASSKEY_AUTH_TRANSPORT_INTERNAL PasskeyAuthTransport = "internal"
)

type (
	PasskeyCredential struct {
		ID         string                 `json:"id"`
		UserId     int                    `json:"user_id"`
		PublicKey  []byte                 `json:"public_key"`
		CreatedAt  time.Time              `json:"created_at"`
		LastUsedAt time.Time              `json:"last_used_at"`
		BackedUp   bool                   `json:"backed_up"`
		Transports []PasskeyAuthTransport `json:"transports"`
	}

	PasskeyCredentialParam struct {
		// Type should be a string representing valid credential type
		Type string
		// Alg is an identifier with restricted set of allowed values (enum)
		Alg int
	}

	// session created during start of passkey registration process
	// required for further verification
	PasskeyRegistrationSession struct {
		UserId     int
		Challenge  string
		CredParams []PasskeyCredentialParam
	}
)