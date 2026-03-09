// Package envvalidator provides struct-based environment variable validation for Go.
package envvalidator

import (
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// ValidationError contains all validation errors collected during parsing.
type ValidationError struct {
	Errors []string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%d validation error(s):\n  - %s", len(e.Errors), strings.Join(e.Errors, "\n  - "))
}

// Validate populates the given struct pointer from environment variables.
// Struct fields are configured via the `env` tag:
//
//	type Config struct {
//	    Port     int    `env:"PORT,default=3000"`
//	    Database string `env:"DATABASE_URL,required"`
//	    Debug    bool   `env:"DEBUG"`
//	}
func Validate(dst any) error {
	return ValidateFrom(dst, nil)
}

// ValidateFrom populates the struct from the given source map instead of os.Getenv.
func ValidateFrom(dst any, source map[string]string) error {
	v := reflect.ValueOf(dst)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("envvalidator: dst must be a pointer to a struct")
	}

	v = v.Elem()
	t := v.Type()
	var errs []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("env")
		if tag == "" {
			continue
		}

		name, opts := parseTag(tag)
		raw := getEnv(name, source)

		if raw == "" {
			if opts.defaultVal != "" {
				raw = opts.defaultVal
			} else if opts.required {
				errs = append(errs, fmt.Sprintf("missing required variable: %s", name))
				continue
			} else {
				continue
			}
		}

		if len(opts.choices) > 0 && !contains(opts.choices, raw) {
			errs = append(errs, fmt.Sprintf("%s must be one of [%s], got '%s'", name, strings.Join(opts.choices, ", "), raw))
			continue
		}

		if err := setField(v.Field(i), raw, name); err != nil {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) > 0 {
		return &ValidationError{Errors: errs}
	}
	return nil
}

type tagOpts struct {
	required   bool
	defaultVal string
	choices    []string
}

func parseTag(tag string) (string, tagOpts) {
	parts := strings.Split(tag, ",")
	name := parts[0]
	opts := tagOpts{}

	for _, p := range parts[1:] {
		p = strings.TrimSpace(p)
		switch {
		case p == "required":
			opts.required = true
		case strings.HasPrefix(p, "default="):
			opts.defaultVal = strings.TrimPrefix(p, "default=")
		case strings.HasPrefix(p, "choices="):
			opts.choices = strings.Split(strings.TrimPrefix(p, "choices="), "|")
		}
	}

	return name, opts
}

func getEnv(name string, source map[string]string) string {
	if source != nil {
		return source[name]
	}
	return os.Getenv(name)
}

func contains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

func setField(field reflect.Value, raw string, name string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(raw)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			d, err := time.ParseDuration(raw)
			if err != nil {
				return fmt.Errorf("%s: invalid duration '%s'", name, raw)
			}
			field.Set(reflect.ValueOf(d))
			return nil
		}
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return fmt.Errorf("%s: cannot convert '%s' to int", name, raw)
		}
		field.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return fmt.Errorf("%s: cannot convert '%s' to uint", name, raw)
		}
		field.SetUint(n)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return fmt.Errorf("%s: cannot convert '%s' to float", name, raw)
		}
		field.SetFloat(f)
	case reflect.Bool:
		b, err := strconv.ParseBool(raw)
		if err != nil {
			return fmt.Errorf("%s: cannot convert '%s' to bool", name, raw)
		}
		field.SetBool(b)
	default:
		// Check for url.URL
		if field.Type() == reflect.TypeOf(url.URL{}) {
			u, err := url.Parse(raw)
			if err != nil {
				return fmt.Errorf("%s: invalid URL '%s'", name, raw)
			}
			field.Set(reflect.ValueOf(*u))
			return nil
		}
		return fmt.Errorf("%s: unsupported type %s", name, field.Type())
	}
	return nil
}
