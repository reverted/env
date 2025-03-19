package env

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// Tag represents the parsed `env` tag
type Tag struct {
	Env      string
	Optional bool
	Default  string
}

// Parse populates a struct with values from environment variables based on struct tags
func Parse(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("input must be a non-nil pointer")
	}

	return parseStruct(rv.Elem())
}

// parseStruct handles parsing of struct fields
func parseStruct(rv reflect.Value) error {
	if rv.Kind() != reflect.Struct {
		return errors.New("value must be a struct")
	}

	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		fieldValue := rv.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Handle embedded or nested structs
		if field.Type.Kind() == reflect.Struct {
			if err := parseStruct(fieldValue); err != nil {
				return err
			}
			continue
		}

		// Handle pointer to struct
		if field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct && !fieldValue.IsNil() {
			if err := parseStruct(fieldValue.Elem()); err != nil {
				return err
			}
			continue
		}

		tag, ok, err := parseTag(field.Tag)
		if err != nil {
			return fmt.Errorf("invalid tag format for field %s: %w", field.Name, err)
		}
		if !ok {
			continue // Field doesn't have an env tag, skip it
		}

		value, exists := os.LookupEnv(tag.Env)
		if !exists {
			if tag.Optional {
				continue
			}
			if tag.Default != "" {
				value = tag.Default
			} else {
				return fmt.Errorf("required environment variable %s not set", tag.Env)
			}
		}

		if err := setField(fieldValue, value); err != nil {
			return fmt.Errorf("failed to set field %s: %w", field.Name, err)
		}
	}

	return nil
}

// parseTag parses the struct tag to extract environment variable configuration
func parseTag(tag reflect.StructTag) (Tag, bool, error) {
	envTag, ok := tag.Lookup("env")
	if !ok {
		return Tag{}, false, nil
	}
	if envTag == "" {
		return Tag{}, false, errors.New("env tag must not be empty")
	}

	parts := strings.Split(envTag, ",")
	if parts[0] == "" {
		return Tag{}, false, errors.New("env tag must have a name")
	}

	result := Tag{
		Env: parts[0],
	}

	for _, part := range parts[1:] {
		if part == "optional" {
			result.Optional = true
		} else if part == "default" {
			// Special handling for "default" without a value
			return Tag{}, false, errors.New("default tag must have a value")
		} else if strings.HasPrefix(part, "default=") {
			if len(part) <= 8 { // 8 is len("default=")
				return Tag{}, false, errors.New("default tag must have a value")
			}
			result.Default = part[8:]
		} else {
			return Tag{}, false, fmt.Errorf("unknown tag option: %s", part)
		}
	}

	return result, true, nil
}

// setField sets the appropriate value to the struct field
func setField(field reflect.Value, value string) error {
	if !field.CanSet() {
		return fmt.Errorf("cannot set field value")
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("cannot parse %s as int: %w", value, err)
		}
		field.SetInt(intValue)
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("cannot parse %s as bool: %w", value, err)
		}
		field.SetBool(boolValue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("cannot parse %s as float: %w", value, err)
		}
		field.SetFloat(floatValue)
	case reflect.Slice:
		sliceValue, err := parseSlice(value, field.Type().Elem())
		if err != nil {
			return err
		}
		field.Set(reflect.ValueOf(sliceValue))
	default:
		return fmt.Errorf("unsupported field type %s", field.Type().String())
	}
	return nil
}

// parseSlice parses a comma-separated string into a slice of the specified type
func parseSlice(value string, elemType reflect.Type) (interface{}, error) {
	if value == "" {
		return reflect.MakeSlice(reflect.SliceOf(elemType), 0, 0).Interface(), nil
	}

	elements := strings.Split(value, ",")

	switch elemType.Kind() {
	case reflect.String:
		return elements, nil
	case reflect.Int:
		result := make([]int, len(elements))
		for i, el := range elements {
			intValue, err := strconv.Atoi(el)
			if err != nil {
				return nil, fmt.Errorf("cannot parse %s as int: %w", el, err)
			}
			result[i] = intValue
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported slice element type %s", elemType.String())
	}
}
