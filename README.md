# go-env-validator

[![CI](https://github.com/philiprehberger/go-env-validator/actions/workflows/ci.yml/badge.svg)](https://github.com/philiprehberger/go-env-validator/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/philiprehberger/go-env-validator.svg)](https://pkg.go.dev/github.com/philiprehberger/go-env-validator)
[![Last updated](https://img.shields.io/github/last-commit/philiprehberger/go-env-validator)](https://github.com/philiprehberger/go-env-validator/commits/main)

![go-env-validator](https://raw.githubusercontent.com/philiprehberger/go-env-validator/main/package-card.webp)

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

### Slices

Delimited values populate slices of any supported scalar type. The default
delimiter is `,`; override it per field with `delim=`.

```go
import "github.com/philiprehberger/go-env-validator"

type Config struct {
    Hosts []string `env:"HOSTS"`          // "a,b,c"  -> [a b c]
    Ports []int    `env:"PORTS,delim=;"`  // "80;443" -> [80 443]
}

// HOSTS="api.example.com,web.example.com" PORTS="80;443"
var cfg Config
_ = envvalidator.Validate(&cfg)
```

`choices` is validated per element for slices. Because the `env` tag is itself
comma-separated, a slice `default` that contains commas must use a non-comma
`delim` (e.g. `env:"HOSTS,delim=;,default=a;b"`).

### Nested Config

Group related variables into nested structs. A field tagged with `envPrefix`
is recursed into, and the prefix is prepended to every child variable name.
Prefixes compose across nesting levels, and `*struct` fields are allocated
automatically.

```go
import "github.com/philiprehberger/go-env-validator"

type DB struct {
    Host string `env:"HOST,default=localhost"`
    Port int    `env:"PORT,default=5432"`
}

type Config struct {
    Database DB `envPrefix:"DB_"` // reads DB_HOST, DB_PORT
}

// DB_HOST=db.internal DB_PORT=6432
var cfg Config
_ = envvalidator.Validate(&cfg)
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
- Slices of any of the above scalar types (e.g. `[]string`, `[]int`, `[]bool`)

### Tag Options

- `required` — field must be set in environment
- `default=VALUE` — fallback if not set (validated against `choices` if both present)
- `choices=A|B|C` — restrict to specific values (whitespace around `|` is trimmed; validated per element for slices)
- `delim=SEP` — element separator for slice fields (default `,`)
- `envPrefix=PREFIX` — on a nested `struct` or `*struct` field, prepend `PREFIX` to every child variable name

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

## Support

If you find this project useful:

⭐ [Star the repo](https://github.com/philiprehberger/go-env-validator)

🐛 [Report issues](https://github.com/philiprehberger/go-env-validator/issues?q=is%3Aissue+is%3Aopen+label%3Abug)

💡 [Suggest features](https://github.com/philiprehberger/go-env-validator/issues?q=is%3Aissue+is%3Aopen+label%3Aenhancement)

❤️ [Sponsor development](https://github.com/sponsors/philiprehberger)

🌐 [All Open Source Projects](https://philiprehberger.com/open-source-packages)

💻 [GitHub Profile](https://github.com/philiprehberger)

🔗 [LinkedIn Profile](https://www.linkedin.com/in/philiprehberger)

## License

[MIT](LICENSE)
