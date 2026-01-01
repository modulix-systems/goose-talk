package redisrepos

import (
	"errors"
	"fmt"

	"github.com/modulix-systems/goose-talk/internal/gateways/storage"
	"github.com/redis/go-redis/v9"
)

func mapError(err error) error {
	if errors.Is(err, redis.Nil) {
		return storage.ErrNotFound
	}
	return err
}

func prefixOtpByEmail(email string) string {
	return fmt.Sprintf("otp:email:%s", email)
}

func prefixOtpByUserId(userId int) string {
	return fmt.Sprintf("otp:user:%d", userId)
}

func prefixPasskeySession(userId int) string {
	return fmt.Sprintf("passkey_session:%d", userId)
}

func prefixQRLoginToken(value string, clientId string) string {
	return fmt.Sprintf("qrlogin:%s:%s", clientId, value)
}

func prefixQRLoginTokenSearchByClient(clientId string) string {
	return fmt.Sprintf("qrlogin:%s:*", clientId)
}
