package utils

import (
	"reflect"
	"strings"
)

// TODO: Unused but an idea to optimize handlers
// getJsonToDbMap generates a mapping of JSON field names to database column names
// for a given struct model. It uses reflection to inspect the struct's fields and
// extracts the `json` and `db` tags to create the mapping. Only fields with both
// valid `json` and `db` tags (not set to "-") are included in the result.
//
// Parameters:
//   - model: An interface{} representing the struct or pointer to a struct
//     from which to extract field mappings.
//
// Returns:
//   - A map[string]string where the keys are JSON tag names and the values
//     are corresponding database column names. If the input is not a struct
//     or pointer to a struct, an empty map is returned.
//
// Example:
//
//	type User struct {
//	    Name  string `json:"name" db:"user_name"`
//	    Email string `json:"email" db:"user_email"`
//	    Secret string `json:"-" db:"secret_key"`
//	}
//	mappings := getJsonToDbMap(User{})
//	// Result: map[string]string{"name": "user_name", "email": "user_email"}
func getJsonToDbMap(model any) map[string]string {
	mappings := make(map[string]string)
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return mappings
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		dbTag := field.Tag.Get("db")
		if jsonTag != "" && dbTag != "" && jsonTag != "-" && dbTag != "-" {
			jsonTag = strings.Split(jsonTag, ",")[0]
			mappings[jsonTag] = dbTag
		}
	}
	return mappings
}
