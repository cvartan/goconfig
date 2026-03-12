# go-config

`go-config` — a lightweight Go package for accessing and managing application configuration.

Configuration values are loaded from the following sources (in order of priority):

1. Configuration files (JSON, YAML, TOML)
2. Environment variables (OS environment variables take precedence)
3. `.env` file (automatically loaded from the application's working directory)

Automatic binding of configuration values to exported struct fields is supported.

---

## Features

- JSON, YAML, TOML configuration files
- Environment variable support (with default values)
- Automatic `.env` file loading from the working directory
- Binding configuration values to struct fields via struct tags
- Automatic type detection (`string`, `bool`, `int64`, `float64`)
- Minimal and explicit Go-style API

---

## Installation
```bash
go get github.com/cvartan/goconfig
```
---

## Configuration File Requirements

### General

- Supported value types: `string`, `bool`, numeric types (stored as `int64` or `float64`)
- Values of unsupported types are ignored
- Environment variable references can be used as values: "${ENV_VARIABLE:DEFAULT_VALUE}"

The value type is determined either by the environment variable or by the default value.
The default value may be omitted (format `${ENV_VARIABLE}`).

If a parameter's value is a string, variable substitution within the value is allowed:
```json
{
    "url": "{$SRV_PROTOCOL:https}://{$SRV_NAME}:{$SRV_PORT:443}/api"
}
```

### `.env` File Support

The library automatically loads environment variables from a `.env` file in the application's working directory (if present).

Format:
```
# Comment lines start with #
DB_HOST=localhost
DB_PORT=5432
DEBUG=true
```

- OS environment variables take precedence over `.env` file values
- The `.env` file is optional and silently ignored if not found
- Supports `KEY=VALUE` format with optional quotes

### JSON and YAML

- The root element must be an object
- Arrays as root elements are not supported

---

## Usage

### Configuration Parsers

Three parsers are available:

- `JsonConfigurationParser` (for `Format = "json"`)
- `YamlConfigurationParser` (for `Format = "yaml"`)
- `TomlConfigurationParser` (for `Format = "toml"`)

You can also create a custom parser (implement the `Parser` interface) and register it via `RegisterParser(format string, parser Parser)`.
```go
type Parser interface {
    Parse(data []byte) (map[string]any, error)
}
```
---

### Configuration File Location

The configuration file can be specified in one of the following ways:

1. **Default** — `config.json` in the application's working directory
2. **Statically** — via options when creating the configuration
3. **Dynamically** — via application flags:

- `-config.path=/opt/app/config`
- `-config.filename=config.yml`
- `-config.format=yaml`
- `-config.source=/opt/app/config/config.yml` (if this parameter is specified, `config.path` and `config.filename` are ignored)

---

### Configuration Options

The `Options` struct is used to configure the configuration manager:
```go
type Options struct {
    Path     string // Path to the configuration file directory (without trailing separator)
    Filename string // Configuration filename (default: config.json)
    Format   string // Data format in the configuration (default supported formats: json, yaml or toml)
    Source   string // Path to the configuration (may refer to non-file sources with custom readers)
}
```
---

## Binding Structs to Configuration (Bind Method)

Configuration parameter values can be bound to exported struct fields using the `config` struct tag.

Rules:

- Fields must be exported
- Field type must match the configuration value type
- If a parameter is missing from the file, an environment variable is used (if set)

Example:
```go
package main

import (
    "github.com/cvartan/goconfig"
)

type AppConfig struct {
    Name  string `config:"app.name"`
    Debug bool   `config:"app.debug"`
}

func main() {
    var cfg AppConfig

    config := goconfig.NewConfiguration(nil) // Use default options

    config.Bind(&cfg) // Populate struct fields
    config.Apply() // Read configuration from file and apply parameter values to bound struct attributes
}
```
---

## API

| Method | Description |
|--------|-------------|
| `NewConfiguration(format string, options ConfigurationOptions) *Configuration` | Create a configuration manager |
| `Configuration.Bind(object any)` | Bind a struct to the configuration (struct must have attributes tagged with `config`) |
| `Configuration.Apply()` | Load and apply configuration |
| `Configuration.Add(key string, valueType ValueDataType)`| Add new parameter without value |
| `Configuration.Set(key string, value any)` | Set a parameter value (or add it if missing) |
| `Configuration.Get(key string) *Parameter` | Get a configuration parameter |
| `Configuration.GetAll() []*Parameter` | Get all parameters |
| `Configuration.Lookup(filter string) []*Parameter` | Get an array of parameters by prefix |
| `Configuration.GetArrayValues(key string) []any` | Get array parameter values |
| `Configuration.GetIntArray(key string) []int64` | Get values for an integer array parameter |
| `Configuration.GetFloatArray(key string) []float64` | Get values for a floating-point array parameter |
| `Configuration.GetStringArray(key string) []string` | Get string values for a string array parameter |
| `Parameter.Name() string` | Get the parameter name |
| `Parameter.Value() any` | Get the raw parameter value |
| `Parameter.Type reflect.Kind` | Get the generic type of the parameter value |
| `Parameter.Int() int64` | Get the parameter value as int |
| `Parameter.Float() float64` | Get the parameter value as float64 |
| `Parameter.Bool() bool` | Get the parameter value as bool |
| `Parameter.String() string` | Get the parameter value as string |

---

## Limitations

- Struct fields that are slices or arrays are not supported

---

## License

MIT