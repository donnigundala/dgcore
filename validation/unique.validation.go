package validation

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

// unique uniqueValidator Custom validator function for unique fields
func unique(db *gorm.DB) validator.Func {
	return func(fl validator.FieldLevel) bool {
		param := strings.Split(fl.Param(), ":")
		if len(param) < 2 {
			return false // Require at least the table name and one column name.
		}

		tableName := param[0]
		fieldNames := strings.Split(strings.Trim(param[1], "[]"), ".")

		// Get the struct that contains the fields.
		structValue := fl.Parent()

		// Build the WHERE clause dynamically based on the fields.
		//whereClause := make(map[string]interface{})
		query := db.Table(tableName)

		for _, field := range fieldNames {
			//trimField := strings.Trim(field, "[]")
			fieldVal := structValue.FieldByName(camelCase(field))

			if !fieldVal.IsValid() {
				return false
			}

			// Use the dynamic table and where clause for the uniqueness check.
			query = query.Where(fmt.Sprintf("%s = ?", field), strings.TrimSpace(fieldVal.Interface().(string)))
		}

		// Handle updates: If an ID is present, exclude the current record from the uniqueness check.
		idField := structValue.FieldByName("ID")
		if idField.IsValid() && !idField.IsZero() {
			query = query.Where("id <> ?", idField.Interface())
		}

		var count int64
		query = query.Where("deleted_at IS NULL")
		if err := query.Count(&count).Error; err != nil {
			return false
		}

		return count == 0
	}
}
