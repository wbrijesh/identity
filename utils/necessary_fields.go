package utils

import (
	"errors"
	"fmt"
	"reflect"
)

func CheckNeceassaryFieldsExist(v interface{}, requiredFields []string) error {
	val := reflect.ValueOf(v)

	// Ensure we're dealing with a struct
	if val.Kind() != reflect.Struct {
		return errors.New("expected a struct")
	}

	// Loop through the required field names
	for _, fieldName := range requiredFields {
		// Get the field by name
		field := val.FieldByName(fieldName)

		// Check if the field exists
		if !field.IsValid() {
			return fmt.Errorf("field %s does not exist in the struct", fieldName)
		}

		// Only check string fields
		if field.Kind() == reflect.String {
			// Check if the string is empty
			if field.String() == "" {
				return fmt.Errorf("field %s is required and cannot be empty", fieldName)
			}
		}
	}

	return nil
}
