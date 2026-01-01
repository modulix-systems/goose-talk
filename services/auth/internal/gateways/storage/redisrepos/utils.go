package redisrepos

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

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
	return fmt.Sprintf("passkey-sessions:%d", userId)
}

func prefixQRLoginToken(value string, clientId string) string {
	return fmt.Sprintf("qrlogin:%s:%s", clientId, value)
}

func prefixQRLoginTokenSearchByClient(clientId string) string {
	return fmt.Sprintf("qrlogin:%s:*", clientId)
}

func prefixAuthSession(userId int, sessionId string) string {
	return fmt.Sprintf("auth-sessions:%d:%s", userId, sessionId)
}

func extractAuthSessionId(key string) string {
	return strings.Split(key, ":")[2]
}

func extractAuthSessionUserId(key string) int {
	idString := strings.Split(key, ":")[1]
	// Ignore error since we know for sure that key contains a a valid user id
	id, _ := strconv.Atoi(idString)
	return id
}

func prefixAuthSessionSearchByUser(userId int) string {
	return fmt.Sprintf("auth-sessions:%d:*", userId)
}