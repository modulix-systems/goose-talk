package gateways

import "time"

type (
	PasskeyCredentialParam struct {
		// Type should be a string representing valid credential type
		Type string
		// Alg is an identifier with restricted set of allowed values (enum)
		Alg int
	}
	// session created during start of passkey registration process
	// required for further verification
	PasskeyTmpSession struct {
		UserId     []byte
		Challenge  string
		CredParams []PasskeyCredentialParam
	}
	TelegramMsg struct {
		DateSent time.Time
		Text     string
		ChatId   string
	}
)
