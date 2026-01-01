package redisrepos

import (
	"context"
	"testing"

	"github.com/modulix-systems/goose-talk/internal/config"
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

type TestSuite struct {
	*Repositories
	RedisClient *redis.Redis
}

func NewTestSuite(t *testing.T) *TestSuite {
	cfg := config.MustLoad(config.ResolveConfigPath("tests"))
	rdb, err := redis.New(cfg.Redis.Url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := rdb.FlushDB(context.Background()).Err(); err != nil {
			t.Fatal(err)
		}
	})
	repos := New(rdb)

	return &TestSuite{repos, rdb}
}
