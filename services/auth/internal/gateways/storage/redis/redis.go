package redis_repos

import (
	otp_repo "github.com/modulix-systems/goose-talk/internal/gateways/storage/redis/otp"
	passkey_repo "github.com/modulix-systems/goose-talk/internal/gateways/storage/redis/passkey"
	qrlogintokens_repo "github.com/modulix-systems/goose-talk/internal/gateways/storage/redis/qr-login-tokens"
	sessions_repo "github.com/modulix-systems/goose-talk/internal/gateways/storage/redis/sessions"
	"github.com/modulix-systems/goose-talk/pkg/redis"
)

type Repositories struct {
	Otp            *otp_repo.Repository
	Sessions       *sessions_repo.Repository
	QRLoginTokens  *qrlogintokens_repo.Repository
	PasskeySession *passkey_repo.Repository
}

func New(rdb *redis.Redis) *Repositories {
	return &Repositories{
		Otp:            &otp_repo.Repository{rdb},
		Sessions:       &sessions_repo.Repository{rdb},
		QRLoginTokens:  &qrlogintokens_repo.Repository{rdb},
		PasskeySession: &passkey_repo.Repository{rdb},
	}
}
