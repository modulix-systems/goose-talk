package helpers

import (
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/internal/gateways"
)

func MockUser() *entity.User {
	return &entity.User{
		ID:        gofakeit.Number(1, 100000),
		Username:  gofakeit.Username(),
		FirstName: gofakeit.FirstName(),
		LastName:  gofakeit.LastName(),
		CreatedAt: gofakeit.Date(),
		UpdatedAt: gofakeit.Date(),
		BirthDate: gofakeit.DateRange(
			time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Now().AddDate(-18, 0, 0),
		),
		PhotoUrl: gofakeit.URL(),
		AboutMe:  gofakeit.Sentence(10),
		IsActive: true,
		Email:    gofakeit.Email(),
		Password: []byte(RandomPassword()),
		TwoFactorAuth: &entity.TwoFactorAuth{
			Transport: RandomChoose(
				entity.TWO_FA_TELEGRAM, entity.TWO_FA_EMAIL,
				entity.TWO_FA_SMS, entity.TWO_FA_TOTP_APP,
			),
			Enabled:    gofakeit.Bool(),
			Contact:    gofakeit.Email(),
			TotpSecret: []byte(gofakeit.Sentence(3)),
		},
	}
}

func MockUserSession(active bool) *entity.UserSession {
	created := gofakeit.DateRange(time.Now().AddDate(0, -1, 0), time.Now())
	lastSeen := gofakeit.DateRange(created, time.Now())

	var deactivated time.Time
	if !active {
		deactivated = gofakeit.DateRange(lastSeen, time.Now())
	}

	return &entity.UserSession{
		ID:             gofakeit.UUID(),
		UserId:         gofakeit.Number(1, 1000),
		ClientIdentity: MockClientIdentity(),
		LastSeenAt:     lastSeen,
		CreatedAt:      created,
		DeactivatedAt:  deactivated,
	}
}

func MockOTP() *entity.OTP {
	return &entity.OTP{
		Code:      []byte(gofakeit.Numerify("######")),
		UserEmail: gofakeit.Email(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func MockLoginToken(ttl time.Duration) *entity.LoginToken {
	return &entity.LoginToken{
		ClientId:         gofakeit.UUID(),
		Val:              gofakeit.UUID(),
		ClientIdentity:   MockClientIdentity(),
		ClientIdentityId: gofakeit.Number(1, 1000),
		AuthSessionId:    gofakeit.Number(0, 1000),
		AuthSession:      MockUserSession(true),
		ExpiresAt:        time.Now().Add(ttl),
	}
}

func MockClientIdentity() *entity.ClientIdentity {
	return &entity.ClientIdentity{
		IPAddr:     gofakeit.IPv4Address(),
		Location:   gofakeit.City(),
		DeviceInfo: gofakeit.UserAgent(),
	}
}

func MockPasskeySession() *gateways.PasskeyTmpSession {
	return &gateways.PasskeyTmpSession{
		UserId:    []byte(gofakeit.Numerify("###")),
		Challenge: gofakeit.Sentence(10),
		CredParams: []gateways.PasskeyCredentialParam{
			gateways.PasskeyCredentialParam{Type: gofakeit.AppName(), Alg: gofakeit.Number(1, 10)},
			gateways.PasskeyCredentialParam{Type: gofakeit.AppName(), Alg: gofakeit.Number(1, 10)},
		},
	}
}

func MockPasskeyCredential() *entity.PasskeyCredential {
	return &entity.PasskeyCredential{
		ID:        gofakeit.UUID(),
		PublicKey: []byte(gofakeit.UUID()),
		UserId:    gofakeit.Number(1, 1000),
		Transports: []entity.PasskeyAuthTransport{
			RandomChoose(
				entity.PASSKEY_AUTH_TRANSPORT_USB,
				entity.PASSKEY_AUTH_TRANSPORT_BLE,
				entity.PASSKEY_AUTH_TRANSPORT_INTERNAL,
				entity.PASSKEY_AUTH_TRANSPORT_NFC,
			),
		},
	}
}
