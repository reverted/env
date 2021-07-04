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
		if !tag.Optional && value == "" {
			return fmt.Errorf("missing required env : %s", tag.Env)
		}

		switch vField.Kind() {
		case reflect.Slice:
			values := strings.Split(value, ",")
			vField.Set(reflect.ValueOf(values))

		case reflect.String:
			vField.SetString(value)

		case reflect.Int:
			parsed, _ := strconv.ParseInt(value, 10, 64)
			vField.SetInt(parsed)

		}
	}

	return nil
}

type Tag struct {
	Env      string
	Optional bool
}

func parseTag(tag reflect.StructTag) (Tag, bool, error) {
	raw, ok := tag.Lookup(TagName)
	if !ok {
		return Tag{}, false, nil
	}

	parts := strings.Split(raw, ",")
	if len(parts) == 0 {
		return Tag{}, false, fmt.Errorf("empty tag '%s'", TagName)
	}

	var optional bool
	for _, value := range parts[1:] {
		if value == "optional" {
			optional = true
		}
	}

	return Tag{
		Env:      parts[0],
		Optional: optional,
	}, true, nil
}
