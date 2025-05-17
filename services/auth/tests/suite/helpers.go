package suite

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/brianvoe/gofakeit/v7"
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

func ChooseRandom[T any](collection ...T) T {
	return collection[rand.Intn(len(collection))]
}

func RandomPassword() string {
	return gofakeit.Password(true, true, true, false, false, 8)
}
