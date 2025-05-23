// Package repo implements application outer layer logic. Each logic group in own file.
package gateways

import (
	"context"
	"errors"
	"time"

	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/internal/schemas"
)

var (
	ErrExpiredToken = errors.New("Expired token")
)

//go:generate mockgen -source=contracts.go -destination=../../tests/mocks/mocks_gateways.go -package=mocks
type (
	UsersRepo interface {
		Insert(ctx context.Context, user *entity.User) (*entity.User, error)
		CheckExistsWithEmail(ctx context.Context, email string) (bool, error)
		GetByLogin(ctx context.Context, login string) (*entity.User, error)
		GetByID(ctx context.Context, id int) (*entity.User, error)
		UpdateIsActiveById(ctx context.Context, userId int, isActive bool) (*entity.User, error)
	}
	UserSessionsRepo interface {
		Insert(ctx context.Context, session *entity.UserSession) (*entity.UserSession, error)
		Delete(ctx context.Context, ip string) error
		GetByToken(ctx context.Context, token string) (*entity.UserSession, error)
		GetAllForUser(
			ctx context.Context,
			userId int,
			activeOnly bool,
		) ([]entity.UserSession, error)
		UpdateById(
			ctx context.Context,
			sessionId int,
			payload *schemas.SessionUpdatePayload,
		) (*entity.UserSession, error)
		UpdateForUserById(
			ctx context.Context,
			userId int,
			sessionId int,
			deactivatedAt time.Time,
		) error
		GetByParamsMatch(
			ctx context.Context,
			ip string,
			deviceInfo string,
			userId int,
		) (*entity.UserSession, error)
	}
	OtpRepo interface {
		GetByEmail(ctx context.Context, email string) (*entity.OTP, error)
		GetByUserId(ctx context.Context, userId int) (*entity.OTP, error)
		DeleteByEmail(ctx context.Context, email string) error
		InsertOrUpdateCode(ctx context.Context, otp *entity.OTP) error
	}
	LoginTokenRepo interface {
		Insert(ctx context.Context, token *entity.LoginToken) (*entity.LoginToken, error)
		GetBySessionId(ctx context.Context, sessionId string) (*entity.LoginToken, error)
		DeleteAllForSessionId(ctx context.Context, sessionId string) error
	}
	TwoFactorAuthRepo interface {
		Insert(ctx context.Context, ent *entity.TwoFactorAuth) (*entity.TwoFactorAuth, error)
		UpdateContactForUser(ctx context.Context, userId int, contact string) error
	}
	AuthTokenProvider interface {
		NewToken(expires time.Duration, claims map[string]any) (string, error)
		ParseClaimsFromToken(token string) (map[string]any, error)
	}
	SecurityProvider interface {
		GenerateOTPCode() string
		GenerateTOTPEnrollUrlWithSecret(accName string) (string, string)
		ValidateTOTP(code string, secret string) bool
		HashPassword(password string) ([]byte, error)
		ComparePasswords(hashed []byte, plain string) (bool, error)
		EncryptSymmetric(plaintext string) ([]byte, error)
		DecryptSymmetric(encrypted []byte) (string, error)
		GenerateSecretTokenUrlSafe(len int) string
	}
	NotificationsService interface {
		SendSignUpConfirmationEmail(ctx context.Context, to string, otp string) error
		SendGreetingEmail(ctx context.Context, to string, name string) error
		Send2FAEmail(ctx context.Context, to string, otp string) error
		SendAccDeactivationEmail(ctx context.Context, to string) error
		SendSignInNewDeviceEmail(ctx context.Context, to string, newSession *entity.UserSession) error
	}
	TelegramMsg struct {
		DateSent time.Time
		Text     string
		ChatId   string
	}
	TelegramBotAPI interface {
		SendTextMsg(ctx context.Context, chatId string, text string) error
		GetStartLinkWithCode(code string) string
		GetLatestMsg(ctx context.Context) (*TelegramMsg, error)
	}
	GeoIPApi interface {
		GetLocationByIP(ip string) (string, error)
	}
)
