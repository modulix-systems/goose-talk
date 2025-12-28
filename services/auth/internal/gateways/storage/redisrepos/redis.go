package redisrepos

import (
	"github.com/modulix-systems/goose-talk/pkg/redis"
)

type Repositories struct {
	Otp            *OtpRepo
	Sessions       *SessionsRepo
	QRLoginTokens  *QRLoginTokensRepo
	PasskeySession *PasskeySessionsRepo
}

func New(rdb *redis.Redis) *Repositories {
	return &Repositories{
		Otp:            &OtpRepo{rdb},
		Sessions:       &SessionsRepo{rdb},
		QRLoginTokens:  &QRLoginTokensRepo{rdb},
		PasskeySession: &PasskeySessionsRepo{rdb},
	}
}
