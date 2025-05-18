package suite

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/modulix-systems/goose-talk/internal/entity"
)

func MockDuration(s string) time.Duration {
	if s == "" {
		s = fmt.Sprintf("%dh", gofakeit.Number(1, 100))
	}
	duration, err := time.ParseDuration(s)
	if err != nil {
		panic(err)
	}
	return duration
}

func RandomChoose[T any](collection ...T) T {
	return collection[rand.Intn(len(collection))]
}

func RandomPassword() string {
	return gofakeit.Password(true, true, true, false, false, 8)
}

func MockUser() *entity.User {
	return &entity.User{
		ID:         gofakeit.Number(1, 100000),
		Username:   gofakeit.Username(),
		FirstName:  gofakeit.FirstName(),
		LastName:   gofakeit.LastName(),
		CreatedAt:  gofakeit.Date(),
		UpdatedAt:  gofakeit.Date(),
		LastSeenAt: gofakeit.Date(),
		BirthDate:  gofakeit.DateRange(time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC), time.Now().AddDate(-18, 0, 0)),
		PhotoUrl:   gofakeit.URL(),
		AboutMe:    gofakeit.Sentence(10),
		IsActive:   true,
		Email:      gofakeit.Email(),
		Password:   []byte(RandomPassword()),
		TwoFactorAuth: &entity.TwoFactorAuth{
			DeliveryMethod: RandomChoose(
				entity.TWO_FA_TELEGRAM, entity.TWO_FA_EMAIL,
				entity.TWO_FA_SMS, entity.TWO_FA_TOTP_APP,
			),
			Enabled:   gofakeit.Bool(),
			Contact:   gofakeit.Email(),
			OtpSecret: gofakeit.Sentence(3),
		},
	}
}

func MockOTP() *entity.OTP {
	return &entity.OTP{
		Code: []byte(gofakeit.Numerify("######")), UserEmail: gofakeit.Email(),
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
}
