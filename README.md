# go-env-validator

Struct-based environment variable validation with batch error reporting for Go.

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

- `string`, `int`, `int64`, `uint`, `float64`, `bool`
- `time.Duration` (e.g., `"30s"`, `"5m"`)
- `url.URL`

## License

MIT
