package commons

import (
	"encore.dev/beta/errs"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func Validate(s interface{}) error {
	err := validate.Struct(s)
	if err != nil {
		return &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "Invalid argument",
			Details: toValidationDetails(err),
		}
	}

	return nil
}

func toValidationDetails(vErr error) errs.ErrDetails {
	var fErrs []fieldDetails
	for _, err := range vErr.(validator.ValidationErrors) {
		fErrs = append(fErrs, fieldDetails{
			Field: err.Field(),
			Err:   err.Tag(),
		})
	}

	return validationDetails{
		Errors: fErrs,
	}
}

type validationDetails struct {
	Errors []fieldDetails `json:"errors"`
}

type fieldDetails struct {
	Field string `json:"field"`
	Err   string `json:"err"`
}

func (validationDetails) ErrDetails() {}
