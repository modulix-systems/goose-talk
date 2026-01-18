package helpers

import (
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/modulix-systems/goose-talk/internal/entity"
)

func MockUser() *entity.User {
	userId := gofakeit.Number(1, 100000)
	return &entity.User{
		Id:        userId,
		Username:  gofakeit.Username(),
		FirstName: gofakeit.FirstName(),
		LastName:  gofakeit.LastName(),
		CreatedAt: gofakeit.Date(),
		UpdatedAt: gofakeit.Date(),
		BirthDate: gofakeit.DateRange(
			time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Now().AddDate(-18, 0, 0),
		),
		PhotoUrl:   gofakeit.URL(),
		AboutMe:    gofakeit.Sentence(10),
		IsActive:   true,
		Email:      gofakeit.Email(),
		Password:   []byte(RandomPassword()),
		PrivateKey: gofakeit.BitcoinPrivateKey(),
		TwoFactorAuth: &entity.TwoFactorAuth{
			UserId: userId,
			Method: RandomChoose(
				entity.TWO_FA_TELEGRAM, entity.TWO_FA_EMAIL,
				entity.TWO_FA_SMS, entity.TWO_FA_TOTP_APP,
			),
			Enabled:    gofakeit.Bool(),
			Contact:    gofakeit.Email(),
			TotpSecret: []byte(gofakeit.Sentence(3)),
		},
	}
}

func MockAuthSession() *entity.AuthSession {
	return &entity.AuthSession{
		Id:         gofakeit.UUID(),
		UserId:     gofakeit.Number(1, 1000),
		IpAddr:     gofakeit.IPv4Address(),
		Location:   gofakeit.City(),
		DeviceInfo: gofakeit.UserAgent(),
	}
}

func MockOTP() *entity.OTP {
	return &entity.OTP{
		Code:      []byte(gofakeit.Numerify("######")),
		UserEmail: gofakeit.Email(),
	}
}

func MockLoginToken() *entity.QRCodeLoginToken {
	return &entity.QRCodeLoginToken{
		ClientId:   gofakeit.UUID(),
		Value:      gofakeit.UUID(),
		IpAddr:     gofakeit.IPv4Address(),
		DeviceInfo: gofakeit.UserAgent(),
	}
}

func MockPasskeySession() *entity.PasskeyRegistrationSession {
	return &entity.PasskeyRegistrationSession{
		UserId:    gofakeit.Number(1, 1000),
		Challenge: gofakeit.Sentence(10),
		CredParams: []entity.PasskeyCredentialParam{
			{Type: gofakeit.AppName(), Alg: gofakeit.Number(1, 10)},
			{Type: gofakeit.AppName(), Alg: gofakeit.Number(1, 10)},
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
