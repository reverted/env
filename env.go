package env

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

const TagName = "env"

func Parse(obj interface{}) error {

	t := reflect.TypeOf(obj)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	v := reflect.ValueOf(obj)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for i := 0; i < t.NumField(); i++ {

		tField := t.Field(i)
		vField := v.Field(i)

		if !vField.CanSet() {
			continue
		}

		tag, ok, err := parseTag(tField.Tag)
		if err != nil {
			return fmt.Errorf("error parsing field '%s' : %w", tField.Name, err)
		}

		if !ok {
			continue
		}

		value := os.Getenv(tag.Env)

		if value == "" {
			value = tag.Default
		}

		if value == "" && !tag.Optional {
			return fmt.Errorf("missing required env : %s", tag.Env)
		}

		switch vField.Kind() {
		case reflect.Slice:
			values, err := parseSlice(value, vField.Type().Elem())
			if err != nil {
				return fmt.Errorf("error parsing slice field '%s' : %w", tField.Name, err)
			}
			vField.Set(reflect.ValueOf(values))

		case reflect.String:
			vField.SetString(value)

		case reflect.Int:
			parsed, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return fmt.Errorf("error parsing int field '%s' : %w", tField.Name, err)
			}
			vField.SetInt(parsed)

		}
	}

	return nil
}

func parseSlice(value string, elemType reflect.Type) (interface{}, error) {
	switch elemType.Kind() {
	case reflect.String:
		return strings.Split(value, ","), nil

	case reflect.Int:
		values := strings.Split(value, ",")
		res := make([]int, len(values))
		for i, v := range values {
			var err error
			if res[i], err = strconv.Atoi(v); err != nil {
				return nil, fmt.Errorf("error parsing int slice : %w", err)
			}
		}
		return res, nil
	}

	return nil, fmt.Errorf("unsupported slice type '%s'", elemType.Kind())
}

type Tag struct {
	Env      string
	Optional bool
	Default  string
}

func parseTag(tag reflect.StructTag) (Tag, bool, error) {
	raw, ok := tag.Lookup(TagName)
	if !ok {
		return Tag{}, false, nil
	}

	parts := strings.Split(raw, ",")
	if len(parts) == 0 || parts[0] == "" {
		return Tag{}, false, fmt.Errorf("empty tag '%s'", TagName)
	}

	var opt bool
	var def string
	for _, value := range parts[1:] {
		switch {
		case value == "optional":
			opt = true

		case strings.HasPrefix(value, "default"):
			defParts := strings.SplitN(value, "=", 2)
			if len(defParts) != 2 {
				return Tag{}, false, fmt.Errorf("invalid use of default in tag '%s', expected 'default=value', found '%s'", TagName, value)
			}
			def = defParts[1]
		}
	}

	return Tag{
		Env:      parts[0],
		Optional: opt,
		Default:  def,
	}, true, nil
}
