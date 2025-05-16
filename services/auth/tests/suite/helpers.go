package suite

import (
	"fmt"
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
