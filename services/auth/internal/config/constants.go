package config

import "time"

const (
	TRANSACTION_CTX_KEY                       = "transaction"
	OTP_LENGTH                                = 6
	LOGIN_TOKEN_LENGTH                        = 16
	AUTH_SESSION_TTL_NEED_INCREMENT_THRESHOLD = 30 * time.Minute
	AUTH_SESSION_TTL_ADDEND                   = 15 * time.Minute
)
