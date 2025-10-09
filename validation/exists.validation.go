package validation

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// exists is custom validator function to check if value is already exist
func exists(db *gorm.DB) validator.Func {
	return func(fl validator.FieldLevel) bool {
		param := strings.Split(fl.Param(), ".")
		if len(param) != 2 {
			return false
		}

		tableName := param[0]
		fieldName := param[1]
		value := fl.Field().Interface()

		var count int64
		db.Table(tableName).Where(fieldName+" = ?", value).Count(&count)

		return count > 0
	}
}
