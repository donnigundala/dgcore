package validation

import (
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type Validator struct {
	validate *validator.Validate
}

type IValidator interface {
	ValidateStruct(i interface{}) error
}

func NewValidator(db *gorm.DB) *Validator {
	v := validator.New()

	// register custom validator
	_ = v.RegisterValidation("unique", unique(db))
	_ = v.RegisterValidation("exists", exists(db))

	return &Validator{validate: v}
}

func (v *Validator) ValidateStruct(i interface{}) error {
	err := v.validate.Struct(i)
	if err != nil {
		return ParseValidationErrors(err, i)
	}
	return nil
}
