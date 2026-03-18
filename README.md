# go-env-validator

[![CI](https://github.com/philiprehberger/go-env-validator/actions/workflows/ci.yml/badge.svg)](https://github.com/philiprehberger/go-env-validator/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/philiprehberger/go-env-validator.svg)](https://pkg.go.dev/github.com/philiprehberger/go-env-validator)
[![License](https://img.shields.io/github/license/philiprehberger/go-env-validator)](LICENSE)

Struct-based environment variable validation with batch error reporting for Go

## Installation

```bash
go get github.com/philiprehberger/go-env-validator
```

## Usage

### Define a Config Struct

```go
import "github.com/philiprehberger/go-env-validator"

type Config struct {
    Port     int           `env:"PORT,default=3000"`
    Database string        `env:"DATABASE_URL,required"`
    Debug    bool          `env:"DEBUG,default=false"`
    Env      string        `env:"APP_ENV,required,choices=development|staging|production"`
    Timeout  time.Duration `env:"TIMEOUT,default=30s"`
}
```

### Validate from Environment

```go
var cfg Config
if err := envvalidator.Validate(&cfg); err != nil {
    log.Fatal(err)
}
fmt.Println(cfg.Port) // 3000
```

### Validate from Map (Testing)

```go
var cfg Config
err := envvalidator.ValidateFrom(&cfg, map[string]string{
    "DATABASE_URL": "postgres://localhost/mydb",
    "APP_ENV":      "development",
})
```

### Batch Error Reporting

```go
if err := envvalidator.Validate(&cfg); err != nil {
    var ve *envvalidator.ValidationError
    if errors.As(err, &ve) {
        for _, e := range ve.Errors {
            fmt.Println(e)
        }
    }
}
```

### Supported Types

- `string`
- `int`, `int8`, `int16`, `int32`, `int64`
- `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- `float32`, `float64`
- `bool` — accepts `true`, `false`, `1`, `0`, `t`, `f`, `T`, `F`, `TRUE`, `FALSE`
- `time.Duration` — Go duration strings (e.g., `"30s"`, `"5m"`, `"1h30m"`)
- `url.URL` — parsed via `url.Parse`
- Any type implementing `encoding.TextUnmarshaler`

### Tag Options

- `required` — field must be set in environment
- `default=VALUE` — fallback if not set (validated against `choices` if both present)
- `choices=A|B|C` — restrict to specific values (whitespace around `|` is trimmed)

## API

| Function / Method | Description |
|---|---|
| `Validate(dst any) error` | Populate struct from environment variables and validate |
| `ValidateFrom(dst any, source map[string]string) error` | Populate struct from a map instead of os.Getenv |
| `ValidationError` | Error type containing all validation errors |
| `(*ValidationError) Error() string` | Format all collected errors as a single string |

## Development

```bash
go test ./...
go vet ./...
```

## License

MIT
