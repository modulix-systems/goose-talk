package validator

import (
	"fmt"
	"reflect"
	"strings"

	govalidator "github.com/go-playground/validator/v10"
	"github.com/modulix-systems/goose-talk/internal/utils"
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
	value := reflect.Indirect(reflect.ValueOf(s)).Type()
	err := v.baseValidaotr.Struct(s)
	if err == nil {
		return nil
	}

	validationErrs := err.(govalidator.ValidationErrors)

	for _, err := range validationErrs {
		structField, _ := value.FieldByName(err.StructField())
		var reason string
		var errorMsg string

		switch err.Tag() {
		case "required":
			errorMsg = "Field is required"
			reason = "MISSING_VALUE"
		case "email":
			errorMsg = "Enter a valid email"
			reason = "INVALID_EMAIL"
		case "ip":
			errorMsg = "Enter a valid IP address"
			reason = "INVALID_IP"
		case "max":
			errorMsg = fmt.Sprintf("The maximum value is %s", err.Param())
			reason = "TOO_LARGE"
		case "min":
			errorMsg = fmt.Sprintf("The minimum value is %s", err.Param())
			reason = "TOO_SMALL"
		case "gte":
			errorMsg = fmt.Sprintf("Value should be greater than or equal to %s", err.Param())
			reason = "TOO_SMALL"
		case "gt":
			errorMsg = fmt.Sprintf("Value should be greater than %s", err.Param())
			reason = "TOO_SMALL"
		case "lte":
			errorMsg = fmt.Sprintf("Value should be less than or equal to %s", err.Param())
			reason = "TOO_LARGE"
		case "lt":
			errorMsg = fmt.Sprintf("Value should be less than %s", err.Param())
			reason = "TOO_LARGE"
		case "eqfield", "eq":
			errorMsg = fmt.Sprintf("Value should be equal to %s", err.Param())
			reason = "NOT_EQUAL"
		case "nefield", "ne":
			errorMsg = fmt.Sprintf("Value should not be equal to %s", err.Param())
			reason = "IS_EQUAL"
		case "oneof":
			errorMsg = fmt.Sprintf("Value should be one of: %s", err.Param())
			reason = "NOT_ONE_OF"
		case "nooneof":
			errorMsg = fmt.Sprintf("Value should not be one of: %s", err.Param())
			reason = "IS_ONE_OF"
		case "len":
			errorMsg = fmt.Sprintf("Length should be equal to %s", err.Param())
		case "http_url":
		case "https_url":
		case "url":
			errorMsg = "Enter a valid url"
		case "unique":
			errorMsg = "Enter non-repitive values"
		default:
			errorMsg = "This field is invalid"
			reason = strings.ToUpper(err.Tag()) + "_VIOLATION"
		}
		if tagErrorMsg := structField.Tag.Get("errorMsg"); tagErrorMsg != "" {
			errorMsg = tagErrorMsg
		}

		fieldName := utils.ToSnakeCase(err.Field())
		v.addError(fieldName, errorMsg, reason)
	}

	return v.Errors
}

func (v *Validator) Check(ok bool, field, description string, reason CustomErrorReason) {
	if !ok {
		v.addError(field, description, string(reason))
	}
}
