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
- Support for custom readers and parsers

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
- Environment variable references can be used as values: `${ENV_VARIABLE:DEFAULT_VALUE}`

The value type is determined either by the environment variable or by the default value.
The default value may be omitted (format `${ENV_VARIABLE}`).

If a parameter's value is a string, variable substitution within the value is allowed:
```json
{
    "url": "${SRV_PROTOCOL:https}://${SRV_NAME}:${SRV_PORT:443}/api"
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

### Configuration Options

The `Options` struct is used to configure the configuration manager:
```go
type Options struct {
    Path          string          // Path to catalog with the configuration file
    Source        string          // Path to the configuration (may refer to non-file sources with custom readers)
    Filename      string          // Filename of the configuration file
    Format        string          // Data format in the configuration (default supported formats: json, yaml or toml)
    Reader        Reader          // Custom configuration reader (see Reader interface)
    Parser        Parser          // Custom configuration parser (see Parser interface)
}
```

When `nil` is passed to `NewConfiguration`, the following defaults are used:
- `Format`: `json`
- `Path`: current working directory
- `Filename`: `config.json`
- `Source`: combination of `Path` and `Filename`

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

### Configuration Parsers

Three parsers are available (registered by default):

- JSON parser (for `Format = "json"`)
- YAML parser (for `Format = "yaml"`)
- TOML parser (for `Format = "toml"`)

You can also create a custom parser (implement the `Parser` interface) and register it via `RegisterParser(format string, parser Parser)`.

---

## Exported Types and Interfaces

### Configuration

The main type for working with application configuration.

```go
type Configuration struct {
    // Contains methods for accessing configuration parameters
}
```

#### Creating a Configuration

```go
func NewConfiguration(options *Options) *Configuration
```

Creates a new configuration manager with the specified options. Pass `nil` to use default options.

#### Methods

| Method | Description |
|--------|-------------|
| `Apply()` | Load and apply the configuration from source |
| `Add(key string, valueType ValueDataType) *Parameter` | Add a new parameter without a value |
| `Set(key string, value any)` | Set a parameter value (or add it if missing) |
| `Get(key string) *Parameter` | Get a configuration parameter |
| `GetAll() []*Parameter` | Get all parameters |
| `Lookup(filter string) []*Parameter` | Get an array of parameters by prefix |
| `Bind(object any)` | Bind a struct to the configuration |
| `GetArrayValues(key string) []any` | Get array parameter values |
| `GetIntArray(key string) []int64` | Get values for an integer array parameter |
| `GetFloatArray(key string) []float64` | Get values for a floating-point array parameter |
| `GetStringArray(key string) []string` | Get string values for a string array parameter |

---

### Parameter

Type for configuration parameter.

```go
type Parameter struct {
    // Contains parameter data with thread-safe access
}
```

#### Methods

| Method | Description |
|--------|-------------|
| `Name() string` | Get the parameter name |
| `Value() any` | Get the raw parameter value |
| `Type() reflect.Kind` | Get the generic type of the parameter value |
| `Int() int64` | Get the parameter value as int64 |
| `Float() float64` | Get the parameter value as float64 |
| `Bool() bool` | Get the parameter value as bool |
| `String() string` | Get the parameter value as string |
| `SetValue(value any)` | Set the parameter value |

---

### StructuredConfiguration

A generic type for working with typed configuration structures.

```go
type StructuredConfiguration struct {
    C *configuration // Configuration
}
```

#### Functions

| Function | Description |
|----------|-------------|
| `NewStructuredConfiguration[T any](options *Options) *T` | Create a typed structured configuration |

#### Methods

| Method | Description |
|--------|-------------|
| `Apply()` | Load and apply the configuration from source |

---

### Reader

Interface for reading configuration data from a source.

```go
type Reader interface {
    Read(source string) ([]byte, error)
}
```

Implement this interface to create custom readers that can read configuration from any source (files, HTTP, databases, etc.).

---

### Parser

Interface for parsing configuration data.

```go
type Parser interface {
    Parse(data []byte) (map[string]any, error)
}
```

Implement this interface to create custom parsers for different data formats. Register custom parsers using `RegisterParser(format string, parser Parser)`.

---

## Binding Structs to Configuration

Configuration parameter values can be bound to exported struct fields using the `config` struct tag.

Rules:

- Fields must be exported
- Field type must match the configuration value type
- If a parameter is missing from the file, an environment variable is used (if set)

There are two ways to bind configuration to structs:

### Method 1: Using Bind

Create a struct with `config` tags, then use the `Bind` and `Apply` methods:

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

### Method 2: Using StructuredConfiguration

Create a struct that embeds `goconfig.StructuredConfiguration`, then initialize it with `NewStructuredConfiguration`:

```go
package main

import (
    "github.com/cvartan/goconfig"
)

type AppConfig struct {
    goconfig.StructuredConfiguration

    Name  string `config:"app.name"`
    Debug bool   `config:"app.debug"`
}

func main() {
    // Initialize the structured configuration with options (or nil for defaults)
    cfg := goconfig.NewStructuredConfiguration[AppConfig](nil)

    cfg.Apply() // Read configuration from file and apply to struct fields

    // Access configuration values directly from struct fields
    // fmt.Println(cfg.Name)
    // fmt.Println(cfg.Debug)
}
```

The `StructuredConfiguration` approach provides:

- Type-safe configuration access through the struct fields
- Direct access to configuration methods via the embedded `StructuredConfiguration` (e.g., `cfg.Apply()`)
- Cleaner code organization by combining configuration structure and behavior

---

## Limitations

- Struct fields that are slices or arrays are not supported

---

## License

MIT