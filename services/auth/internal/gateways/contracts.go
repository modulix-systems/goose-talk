package gateways

import (
	"context"
	"time"

	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/internal/schemas"
)

//go:generate mockgen -source=contracts.go -destination=../../tests/mocks/mocks_gateways.go -package=mocks
type (
	UsersRepo interface {
		Insert(ctx context.Context, user *entity.User) (*entity.User, error)
		CheckExistsWithEmail(ctx context.Context, email string) (bool, error)
		GetByLogin(ctx context.Context, login string) (*entity.User, error)
		GetByID(ctx context.Context, id int) (*entity.User, error)
		GetByIDWithPasskeyCredentials(ctx context.Context, id int) (*entity.User, error)
		UpdateIsActiveById(ctx context.Context, userId int, isActive bool) (*entity.User, error)
		AddPasskeyCredential(ctx context.Context, userId int, cred *entity.PasskeyCredential) error
	}
	UserSessionsRepo interface {
		Insert(ctx context.Context, session *entity.UserSession) (*entity.UserSession, error)
		Delete(ctx context.Context, ip string) error
		GetAllForUser(
			ctx context.Context,
			userId int,
			activeOnly bool,
		) ([]entity.UserSession, error)
		UpdateById(
			ctx context.Context,
			sessionId string,
			payload *schemas.SessionUpdatePayload,
		) (*entity.UserSession, error)
		UpdateForUserById(
			ctx context.Context,
			userId int,
			sessionId string,
			deactivatedAt time.Time,
		) error
		GetByParamsMatch(
			ctx context.Context,
			ip string,
			deviceInfo string,
			userId int,
		) (*entity.UserSession, error)
		GetById(ctx context.Context, sessionId string) (*entity.UserSession, error)
	}
	OtpRepo interface {
		GetByEmail(ctx context.Context, email string) (*entity.OTP, error)
		GetByUserId(ctx context.Context, userId int) (*entity.OTP, error)
		DeleteByEmailOrUserId(ctx context.Context, email string, userId int) error
		InsertOrUpdateCode(ctx context.Context, otp *entity.OTP) error
	}
	LoginTokenRepo interface {
		Insert(ctx context.Context, token *entity.LoginToken) (*entity.LoginToken, error)
		GetByClientId(ctx context.Context, sessionId string) (*entity.LoginToken, error)
		GetByValue(ctx context.Context, val string) (*entity.LoginToken, error)
		DeleteByClientId(ctx context.Context, sessionId string) error
		UpdateAuthSessionByClientId(ctx context.Context, clientId string, authSessionId string) error
	}
	TwoFactorAuthRepo interface {
		Insert(ctx context.Context, ent *entity.TwoFactorAuth) (*entity.TwoFactorAuth, error)
		UpdateContactForUser(ctx context.Context, userId int, contact string) error
	}
	SecurityProvider interface {
		GenerateOTPCode() string
		GenerateTOTPEnrollUrlWithSecret(accName string) (string, string)
		ValidateTOTP(code string, secret string) bool
		HashPassword(password string) ([]byte, error)
		ComparePasswords(hashed []byte, plain string) (bool, error)
		EncryptSymmetric(plaintext string) ([]byte, error)
		DecryptSymmetric(encrypted []byte) (string, error)
		GenerateSecretTokenUrlSafe(entropy int) string
		GenerateSessionId() string
	}
	KeyValueStorage interface {
		Set(key string, value string, expiresIn time.Duration) error
		Get(key string) (string, error)
	}
	WebAuthnRegistrationOptions []byte
	WebAuthnProvider            interface {
		GenerateRegistrationOptions(user *entity.User) (WebAuthnRegistrationOptions, *PasskeyTmpSession, error)
		VerifyRegistrationOptions(userId int, rawCredential []byte, prevSession *PasskeyTmpSession) (*entity.PasskeyCredential, error)
	}
	NotificationsService interface {
		SendSignUpConfirmationEmail(ctx context.Context, to string, otp string) error
		SendGreetingEmail(ctx context.Context, to string, name string) error
		Send2FAEmail(ctx context.Context, to string, otp string) error
		SendAccDeactivationEmail(ctx context.Context, to string) error
		SendSignInNewDeviceEmail(ctx context.Context, to string, newSession *entity.UserSession) error
	}
	TelegramBotAPI interface {
		SendTextMsg(ctx context.Context, chatId string, text string) error
		GetStartLinkWithCode(code string) string
		GetLatestMsg(ctx context.Context) (*TelegramMsg, error)
	}
	GeoIPApi interface {
		GetLocationByIP(ip string) (string, error)
	}
	Transaction interface {
		Commit(ctx context.Context) error
		Rollback(ctx context.Context) error
	}
	TransactionsManager interface {
		StartTransaction(ctx context.Context) (Transaction, error)
	}
)
