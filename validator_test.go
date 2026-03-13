package envvalidator

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"
)

// netAddr implements encoding.TextUnmarshaler for testing
type netAddr struct {
	Host string
	Port string
}

func (a *netAddr) UnmarshalText(text []byte) error {
	parts := strings.SplitN(string(text), ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid address: missing colon")
	}
	a.Host = parts[0]
	a.Port = parts[1]
	return nil
}

// String types

func TestStringField(t *testing.T) {
	type Config struct {
		Name string `env:"NAME"`
	}
	var cfg Config
	err := ValidateFrom(&cfg, map[string]string{"NAME": "alice"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Name != "alice" {
		t.Fatalf("expected 'alice', got %q", cfg.Name)
	}
}

func TestIntField(t *testing.T) {
	type Config struct {
		Port int `env:"PORT"`
	}
	var cfg Config
	err := ValidateFrom(&cfg, map[string]string{"PORT": "8080"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 8080 {
		t.Fatalf("expected 8080, got %d", cfg.Port)
	}
}

func TestInt64Field(t *testing.T) {
	type Config struct {
		Size int64 `env:"SIZE"`
	}
	var cfg Config
	err := ValidateFrom(&cfg, map[string]string{"SIZE": "9999999999"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Size != 9999999999 {
		t.Fatalf("expected 9999999999, got %d", cfg.Size)
	}
}

func TestUintField(t *testing.T) {
	type Config struct {
		Count uint `env:"COUNT"`
	}
	var cfg Config
	err := ValidateFrom(&cfg, map[string]string{"COUNT": "42"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Count != 42 {
		t.Fatalf("expected 42, got %d", cfg.Count)
	}
}

func TestFloat64Field(t *testing.T) {
	type Config struct {
		Rate float64 `env:"RATE"`
	}
	var cfg Config
	err := ValidateFrom(&cfg, map[string]string{"RATE": "3.14"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Rate != 3.14 {
		t.Fatalf("expected 3.14, got %f", cfg.Rate)
	}
}

func TestBoolField(t *testing.T) {
	type Config struct {
		Debug bool `env:"DEBUG"`
	}

	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1", true},
		{"0", false},
		{"TRUE", true},
		{"FALSE", false},
		{"t", true},
		{"f", false},
	}
	for _, tc := range tests {
		var cfg Config
		err := ValidateFrom(&cfg, map[string]string{"DEBUG": tc.input})
		if err != nil {
			t.Fatalf("input %q: unexpected error: %v", tc.input, err)
		}
		if cfg.Debug != tc.expected {
			t.Fatalf("input %q: expected %v, got %v", tc.input, tc.expected, cfg.Debug)
		}
	}
}

func TestDurationField(t *testing.T) {
	type Config struct {
		Timeout time.Duration `env:"TIMEOUT"`
	}
	var cfg Config
	err := ValidateFrom(&cfg, map[string]string{"TIMEOUT": "30s"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Timeout != 30*time.Second {
		t.Fatalf("expected 30s, got %v", cfg.Timeout)
	}
}

func TestURLField(t *testing.T) {
	type Config struct {
		Endpoint url.URL `env:"ENDPOINT"`
	}
	var cfg Config
	err := ValidateFrom(&cfg, map[string]string{"ENDPOINT": "https://example.com/api"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Endpoint.Host != "example.com" {
		t.Fatalf("expected host example.com, got %q", cfg.Endpoint.Host)
	}
}

// Tag options

func TestRequired(t *testing.T) {
	type Config struct {
		DB string `env:"DB,required"`
	}
	var cfg Config
	err := ValidateFrom(&cfg, map[string]string{})

	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if len(ve.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(ve.Errors))
	}
}

func TestDefault(t *testing.T) {
	type Config struct {
		Port int `env:"PORT,default=3000"`
	}
	var cfg Config
	err := ValidateFrom(&cfg, map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 3000 {
		t.Fatalf("expected 3000, got %d", cfg.Port)
	}
}

func TestDefaultOverriddenByValue(t *testing.T) {
	type Config struct {
		Port int `env:"PORT,default=3000"`
	}
	var cfg Config
	err := ValidateFrom(&cfg, map[string]string{"PORT": "8080"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 8080 {
		t.Fatalf("expected 8080, got %d", cfg.Port)
	}
}

func TestChoices(t *testing.T) {
	type Config struct {
		Env string `env:"APP_ENV,choices=dev|staging|prod"`
	}
	var cfg Config
	err := ValidateFrom(&cfg, map[string]string{"APP_ENV": "dev"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Env != "dev" {
		t.Fatalf("expected 'dev', got %q", cfg.Env)
	}
}

func TestChoicesRejected(t *testing.T) {
	type Config struct {
		Env string `env:"APP_ENV,choices=dev|staging|prod"`
	}
	var cfg Config
	err := ValidateFrom(&cfg, map[string]string{"APP_ENV": "invalid"})
	if err == nil {
		t.Fatal("expected error for invalid choice")
	}
}

func TestChoicesWhitespaceTrimmed(t *testing.T) {
	type Config struct {
		Env string `env:"APP_ENV,choices=dev | staging | prod"`
	}
	var cfg Config
	err := ValidateFrom(&cfg, map[string]string{"APP_ENV": "staging"})
	if err != nil {
		t.Fatalf("expected whitespace-trimmed choice to match, got: %v", err)
	}
}

func TestDefaultValidatedAgainstChoices(t *testing.T) {
	type Config struct {
		Env string `env:"APP_ENV,default=invalid,choices=dev|prod"`
	}
	var cfg Config
	err := ValidateFrom(&cfg, map[string]string{})
	if err == nil {
		t.Fatal("expected error for default not in choices")
	}
}

// Batch errors

func TestMultipleValidationErrors(t *testing.T) {
	type Config struct {
		DB   string `env:"DB,required"`
		Port int    `env:"PORT,required"`
	}
	var cfg Config
	err := ValidateFrom(&cfg, map[string]string{})

	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if len(ve.Errors) != 2 {
		t.Fatalf("expected 2 errors, got %d: %v", len(ve.Errors), ve.Errors)
	}
}

func TestValidationErrorMessage(t *testing.T) {
	type Config struct {
		DB string `env:"DB,required"`
	}
	var cfg Config
	err := ValidateFrom(&cfg, map[string]string{})
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	if msg == "" {
		t.Fatal("expected non-empty error message")
	}
}

// Type conversion errors

func TestInvalidInt(t *testing.T) {
	type Config struct {
		Port int `env:"PORT"`
	}
	var cfg Config
	err := ValidateFrom(&cfg, map[string]string{"PORT": "notanumber"})
	if err == nil {
		t.Fatal("expected error for invalid int")
	}
}

func TestInvalidUint(t *testing.T) {
	type Config struct {
		Count uint `env:"COUNT"`
	}
	var cfg Config
	err := ValidateFrom(&cfg, map[string]string{"COUNT": "-5"})
	if err == nil {
		t.Fatal("expected error for negative uint")
	}
}

func TestInvalidFloat(t *testing.T) {
	type Config struct {
		Rate float64 `env:"RATE"`
	}
	var cfg Config
	err := ValidateFrom(&cfg, map[string]string{"RATE": "notafloat"})
	if err == nil {
		t.Fatal("expected error for invalid float")
	}
}

func TestInvalidBool(t *testing.T) {
	type Config struct {
		Debug bool `env:"DEBUG"`
	}
	var cfg Config
	err := ValidateFrom(&cfg, map[string]string{"DEBUG": "notabool"})
	if err == nil {
		t.Fatal("expected error for invalid bool")
	}
}

func TestInvalidDuration(t *testing.T) {
	type Config struct {
		Timeout time.Duration `env:"TIMEOUT"`
	}
	var cfg Config
	err := ValidateFrom(&cfg, map[string]string{"TIMEOUT": "notaduration"})
	if err == nil {
		t.Fatal("expected error for invalid duration")
	}
}

// Edge cases

func TestNoTagSkipped(t *testing.T) {
	type Config struct {
		Internal string
		Name     string `env:"NAME"`
	}
	var cfg Config
	err := ValidateFrom(&cfg, map[string]string{"NAME": "alice"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Internal != "" {
		t.Fatal("expected untagged field to be untouched")
	}
}

func TestNotAPointer(t *testing.T) {
	type Config struct{}
	var cfg Config
	err := ValidateFrom(cfg, nil)
	if err == nil {
		t.Fatal("expected error for non-pointer")
	}
}

func TestNotAStruct(t *testing.T) {
	s := "not a struct"
	err := ValidateFrom(&s, nil)
	if err == nil {
		t.Fatal("expected error for non-struct pointer")
	}
}

func TestEmptySource(t *testing.T) {
	type Config struct {
		Name string `env:"NAME,default=fallback"`
	}
	var cfg Config
	err := ValidateFrom(&cfg, map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Name != "fallback" {
		t.Fatalf("expected 'fallback', got %q", cfg.Name)
	}
}

func TestOptionalFieldLeftZero(t *testing.T) {
	type Config struct {
		Port int `env:"PORT"`
	}
	var cfg Config
	err := ValidateFrom(&cfg, map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 0 {
		t.Fatalf("expected 0 for unset optional, got %d", cfg.Port)
	}
}

func TestUnsupportedType(t *testing.T) {
	type Custom struct{ X int }
	type Config struct {
		C Custom `env:"C"`
	}
	var cfg Config
	err := ValidateFrom(&cfg, map[string]string{"C": "value"})
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
}

func TestInt8Overflow(t *testing.T) {
	type Config struct {
		Small int8 `env:"SMALL"`
	}
	var cfg Config
	err := ValidateFrom(&cfg, map[string]string{"SMALL": "200"})
	if err == nil {
		t.Fatal("expected error for int8 overflow (200 > 127)")
	}
}

func TestUint8Overflow(t *testing.T) {
	type Config struct {
		Small uint8 `env:"SMALL"`
	}
	var cfg Config
	err := ValidateFrom(&cfg, map[string]string{"SMALL": "300"})
	if err == nil {
		t.Fatal("expected error for uint8 overflow (300 > 255)")
	}
}

func TestInt8Valid(t *testing.T) {
	type Config struct {
		Small int8 `env:"SMALL"`
	}
	var cfg Config
	err := ValidateFrom(&cfg, map[string]string{"SMALL": "127"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Small != 127 {
		t.Fatalf("expected 127, got %d", cfg.Small)
	}
}

func TestTextUnmarshaler(t *testing.T) {
	type Config struct {
		Addr netAddr `env:"ADDR"`
	}
	var cfg Config
	err := ValidateFrom(&cfg, map[string]string{"ADDR": "192.168.1.1:8080"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Addr.Host != "192.168.1.1" || cfg.Addr.Port != "8080" {
		t.Fatalf("expected 192.168.1.1:8080, got %s:%s", cfg.Addr.Host, cfg.Addr.Port)
	}
}

func TestTextUnmarshalerError(t *testing.T) {
	type Config struct {
		Addr netAddr `env:"ADDR"`
	}
	var cfg Config
	err := ValidateFrom(&cfg, map[string]string{"ADDR": "no-colon"})
	if err == nil {
		t.Fatal("expected error for invalid TextUnmarshaler input")
	}
}

func TestMultipleFields(t *testing.T) {
	type Config struct {
		Host    string        `env:"HOST,default=localhost"`
		Port    int           `env:"PORT,default=3000"`
		Debug   bool          `env:"DEBUG,default=false"`
		Timeout time.Duration `env:"TIMEOUT,default=5s"`
	}
	var cfg Config
	err := ValidateFrom(&cfg, map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Host != "localhost" || cfg.Port != 3000 || cfg.Debug != false || cfg.Timeout != 5*time.Second {
		t.Fatalf("defaults not applied correctly: %+v", cfg)
	}
}
