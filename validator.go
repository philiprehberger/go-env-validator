// Package envvalidator provides struct-based environment variable validation for Go.
package envvalidator

import (
	"encoding"
	"fmt"
	"math"
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

	var errs []string
	validateStruct(v.Elem(), source, "", &errs)

	if len(errs) > 0 {
		return &ValidationError{Errors: errs}
	}
	return nil
}

// validateStruct walks the fields of v, applying prefix to every env name.
// Nested structs tagged with `envPrefix` are recursed into with the prefix
// extended, allowing related configuration to be grouped.
func validateStruct(v reflect.Value, source map[string]string, prefix string, errs *[]string) {
	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Nested config struct: recurse with an extended prefix.
		if sub, ok := field.Tag.Lookup("envPrefix"); ok {
			nested, ok := structValue(v.Field(i))
			if !ok {
				*errs = append(*errs, fmt.Sprintf("%s: envPrefix requires a struct or *struct field", field.Name))
				continue
			}
			validateStruct(nested, source, prefix+sub, errs)
			continue
		}

		tag := field.Tag.Get("env")
		if tag == "" {
			continue
		}

		name, opts := parseTag(tag)
		name = prefix + name
		raw := getEnv(name, source)

		if raw == "" {
			if opts.defaultVal != "" {
				raw = opts.defaultVal
			} else if opts.required {
				*errs = append(*errs, fmt.Sprintf("missing required variable: %s", name))
				continue
			} else {
				continue
			}
		}

		isSlice := v.Field(i).Kind() == reflect.Slice

		// For scalars, choices are validated against the whole value; for
		// slices they are validated per element inside setField.
		if !isSlice && len(opts.choices) > 0 && !contains(opts.choices, raw) {
			*errs = append(*errs, fmt.Sprintf("%s must be one of [%s], got '%s'", name, strings.Join(opts.choices, ", "), raw))
			continue
		}

		// Validate scalar default values against choices.
		if !isSlice && opts.defaultVal != "" && len(opts.choices) > 0 && !contains(opts.choices, opts.defaultVal) {
			*errs = append(*errs, fmt.Sprintf("%s: default value '%s' is not one of [%s]", name, opts.defaultVal, strings.Join(opts.choices, ", ")))
			continue
		}

		if err := setField(v.Field(i), raw, name, opts); err != nil {
			*errs = append(*errs, err.Error())
		}
	}
}

// structValue returns the addressable struct value for a nested field,
// allocating a *struct if it is nil. The second return is false if the field
// is neither a struct nor a pointer to one.
func structValue(field reflect.Value) (reflect.Value, bool) {
	if field.Kind() == reflect.Ptr {
		if field.Type().Elem().Kind() != reflect.Struct {
			return reflect.Value{}, false
		}
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		return field.Elem(), true
	}
	if field.Kind() == reflect.Struct {
		return field, true
	}
	return reflect.Value{}, false
}

type tagOpts struct {
	required   bool
	defaultVal string
	choices    []string
	delim      string
}

func parseTag(tag string) (string, tagOpts) {
	parts := strings.Split(tag, ",")
	name := parts[0]
	opts := tagOpts{delim: ","}

	for _, p := range parts[1:] {
		p = strings.TrimSpace(p)
		switch {
		case p == "required":
			opts.required = true
		case strings.HasPrefix(p, "default="):
			opts.defaultVal = strings.TrimPrefix(p, "default=")
		case strings.HasPrefix(p, "delim="):
			if d := strings.TrimPrefix(p, "delim="); d != "" {
				opts.delim = d
			}
		case strings.HasPrefix(p, "choices="):
			raw := strings.Split(strings.TrimPrefix(p, "choices="), "|")
			opts.choices = make([]string, len(raw))
			for i, c := range raw {
				opts.choices[i] = strings.TrimSpace(c)
			}
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

// setField dispatches a raw string to the appropriate setter. Slice fields are
// split on opts.delim and each element is converted individually (and checked
// against choices, if any); all other fields go through setScalar.
func setField(field reflect.Value, raw string, name string, opts tagOpts) error {
	if field.Kind() == reflect.Slice {
		parts := strings.Split(raw, opts.delim)
		slice := reflect.MakeSlice(field.Type(), len(parts), len(parts))
		for i, p := range parts {
			p = strings.TrimSpace(p)
			if len(opts.choices) > 0 && !contains(opts.choices, p) {
				return fmt.Errorf("%s must be one of [%s], got '%s'", name, strings.Join(opts.choices, ", "), p)
			}
			if err := setScalar(slice.Index(i), p, name); err != nil {
				return err
			}
		}
		field.Set(slice)
		return nil
	}
	return setScalar(field, raw, name)
}

func setScalar(field reflect.Value, raw string, name string) error {
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
		bitSize := field.Type().Bits()
		n, err := strconv.ParseInt(raw, 10, bitSize)
		if err != nil {
			return fmt.Errorf("%s: cannot convert '%s' to %s", name, raw, field.Type())
		}
		field.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		bitSize := field.Type().Bits()
		n, err := strconv.ParseUint(raw, 10, bitSize)
		if err != nil {
			return fmt.Errorf("%s: cannot convert '%s' to %s", name, raw, field.Type())
		}
		field.SetUint(n)
	case reflect.Float32, reflect.Float64:
		bitSize := 64
		if field.Kind() == reflect.Float32 {
			bitSize = 32
		}
		f, err := strconv.ParseFloat(raw, bitSize)
		if err != nil {
			return fmt.Errorf("%s: cannot convert '%s' to %s", name, raw, field.Type())
		}
		if bitSize == 32 && (f > math.MaxFloat32 || f < -math.MaxFloat32) {
			return fmt.Errorf("%s: value '%s' overflows float32", name, raw)
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
		// Check for encoding.TextUnmarshaler interface
		if field.CanAddr() {
			if tu, ok := field.Addr().Interface().(encoding.TextUnmarshaler); ok {
				if err := tu.UnmarshalText([]byte(raw)); err != nil {
					return fmt.Errorf("%s: failed to unmarshal '%s': %w", name, raw, err)
				}
				return nil
			}
		}
		return fmt.Errorf("%s: unsupported type %s", name, field.Type())
	}
	return nil
}
