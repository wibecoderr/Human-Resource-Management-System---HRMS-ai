package utils

import (
	"fmt"
	"hrms/model"
	"reflect"
	"strings"
)

// ValidateStruct does basic required + type validation using struct tags.
// For a hackathon this is enough; swap in go-playground/validator later if needed.
func ValidateStruct(s interface{}) []model.ValidationError {
	var errs []model.ValidationError

	v := reflect.ValueOf(s)
	t := reflect.TypeOf(s)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		tag := field.Tag.Get("validate")
		if tag == "" {
			continue
		}

		jsonName := field.Tag.Get("json")
		if jsonName == "" {
			jsonName = strings.ToLower(field.Name)
		}

		rules := strings.Split(tag, ",")
		for _, rule := range rules {
			rule = strings.TrimSpace(rule)

			if rule == "required" {
				if value.Kind() == reflect.String && value.String() == "" {
					errs = append(errs, model.ValidationError{
						Field:   jsonName,
						Message: fmt.Sprintf("%s is required", jsonName),
					})
				}
			}

			if strings.HasPrefix(rule, "min=") {
				minLen := 0
				fmt.Sscanf(rule, "min=%d", &minLen)
				if value.Kind() == reflect.String && len(value.String()) < minLen {
					errs = append(errs, model.ValidationError{
						Field:   jsonName,
						Message: fmt.Sprintf("%s must be at least %d characters", jsonName, minLen),
					})
				}
			}

			if rule == "email" {
				if value.Kind() == reflect.String && !strings.Contains(value.String(), "@") {
					errs = append(errs, model.ValidationError{
						Field:   jsonName,
						Message: fmt.Sprintf("%s must be a valid email", jsonName),
					})
				}
			}

			if strings.HasPrefix(rule, "oneof=") {
				allowed := strings.Fields(strings.TrimPrefix(rule, "oneof="))
				val := value.String()
				found := false
				for _, a := range allowed {
					if val == a {
						found = true
						break
					}
				}
				if !found {
					errs = append(errs, model.ValidationError{
						Field:   jsonName,
						Message: fmt.Sprintf("%s must be one of: %s", jsonName, strings.Join(allowed, ", ")),
					})
				}
			}
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}
