package gateways

import (
	"context"
	"time"

	"github.com/modulix-systems/goose-talk/internal/entity"
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
		CreatePasskeyCredential(ctx context.Context, userId int, cred *entity.PasskeyCredential) error
		CreateTwoFa(ctx context.Context, ent *entity.TwoFactorAuth) (*entity.TwoFactorAuth, error)
		UpdateTwoFaContact(ctx context.Context, userId int, contact string) error
	}
	UserSessionsRepo interface {
		CreateWithTTL(ctx context.Context, session *entity.AuthSession, ttl time.Duration) (*entity.AuthSession, error)
		DeleteById(ctx context.Context, id string) error
		GetAllByUserId(ctx context.Context, userId int) ([]entity.AuthSession, error)
		GetByLoginData(ctx context.Context, ip string, deviceInfo string, userId int) (*entity.AuthSession, error)
		GetById(ctx context.Context, id string) (*entity.AuthSession, error)
		UpdateById(ctx context.Context, sessionId string, lastSeenAt time.Time, ttl time.Duration) (*entity.AuthSession, error)
	}
	OtpRepo interface {
		GetByEmail(ctx context.Context, email string) (*entity.OTP, error)
		GetByUserId(ctx context.Context, userId int) (*entity.OTP, error)
		Delete(ctx context.Context, otp *entity.OTP) error
		CreateWithTTL(ctx context.Context, otp *entity.OTP, ttl time.Duration) error
	}
	QRLoginTokenRepo interface {
		CreateWithTTL(ctx context.Context, token *entity.QRCodeLoginToken, ttl time.Duration) (*entity.QRCodeLoginToken, error)
		GetByValue(ctx context.Context, val string) (*entity.QRCodeLoginToken, error)
		DeleteAllByClientId(ctx context.Context, sessionId string) error
		DeleteByValue(ctx context.Context, val string) error
	}
	SecurityProvider interface {
		GenerateOTPCode() string
		GenerateTOTPEnrollUrlWithSecret(accName string) (string, string)
		ValidateTOTP(code string, secret string) bool
		HashPassword(password string) ([]byte, error)
		ComparePasswords(hashed []byte, plain string) error
		EncryptSymmetric(plaintext string, key string) ([]byte, error)
		DecryptSymmetric(encrypted []byte, key string) (string, error)
		GenerateSecretTokenUrlSafe(entropy int) string
		GenerateSessionId() string
		GeneratePrivateKey() string
	}
	PasskeySessionsRepo interface {
		Create(ctx context.Context, session *entity.PasskeyRegistrationSession) error
		GetByUserId(ctx context.Context, userId int) (*entity.PasskeyRegistrationSession, error)
	}
	WebAuthnRegistrationOptions []byte
	WebAuthnProvider            interface {
		GenerateRegistrationOptions(user *entity.User) (WebAuthnRegistrationOptions, *entity.PasskeyRegistrationSession, error)
		VerifyRegistrationOptions(userId int, rawCredential []byte, prevSession *entity.PasskeyRegistrationSession) (*entity.PasskeyCredential, error)
	}
	NotificationsClient interface {
		SendSignUpConfirmationEmail(ctx context.Context, to string, otp string) error
		SendGreetingEmail(ctx context.Context, to string, name string) error
		Send2FAEmail(ctx context.Context, to string, otp string) error
		SendAccDeactivationEmail(ctx context.Context, to string) error
		SendSignInNewDeviceEmail(ctx context.Context, to string, newSession *entity.AuthSession) error
	}
	TelegramBotClient interface {
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
