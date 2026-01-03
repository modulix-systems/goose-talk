package validator

import (
	"strings"

	govalidator "github.com/go-playground/validator/v10"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

type Validator struct {
	baseValidaotr *govalidator.Validate
	Errors        ValidationErrors
}

func New() *Validator {
	return &Validator{
		baseValidaotr: baseValidator,
		Errors:        []*FieldError{},
	}
}

type FieldError = errdetails.BadRequest_FieldViolation
type ValidationErrors = []*FieldError

type CustomErrorReason string

var (
	TOO_YOUNG CustomErrorReason = "TOO_YOUNG"
)

var baseValidator = govalidator.New(govalidator.WithRequiredStructEnabled())

func (v *Validator) addError(field, description, reason string) {
	v.Errors = append(v.Errors, &FieldError{
		Field:       field,
		Description: description,
		Reason:      reason,
	})
}

func (v *Validator) ValidateStruct(s interface{}) ValidationErrors {
	if err := v.baseValidaotr.Struct(s); err != nil {
		validationErrs := err.(govalidator.ValidationErrors)
		for _, err := range validationErrs {
			v.addError(err.Field(), err.Error(), strings.ToUpper(err.Tag())+"_VIOLATION")
		}

		return v.Errors
	}

	return nil
}
func (v *Validator) Check(ok bool, field, description string, reason CustomErrorReason) {
	if !ok {
		v.addError(field, description, string(reason))
	}
}
