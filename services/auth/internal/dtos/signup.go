package dtos

import (
	"time"

	"github.com/modulix-systems/goose-talk/internal/entity"
	"github.com/modulix-systems/goose-talk/pkg/validator"
)

type SignUpRequest struct {
	Username         string `validate:"required"`
	Password         string `validate:"required,min=8"`
	Email            string `validate:"required,email"`
	FirstName        string
	LastName         string
	ConfirmationCode string `validate:"required,len=6"`
	IpAddr           string `validate:"required,ip"`
	DeviceInfo       string `validate:"required"`
	BirthDate        time.Time
	AboutMe          string
}

func (req *SignUpRequest) Validate() validator.ValidationErrors {
	validate := validator.New()
	validate.ValidateStruct(req)
	if !req.BirthDate.IsZero() {
		age := time.Now().Year() - req.BirthDate.Year()
		validate.Check(age >= 18, "birth_date", "You should be at least 18 years old to use this service", validator.TOO_YOUNG)
	}
	return validate.Errors
}

type SignUpResponse struct {
	Session *entity.AuthSession
	User    *entity.User
}
