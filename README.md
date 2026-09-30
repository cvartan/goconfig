# go-config

`go-config` is a lightweight Go package for accessing and managing application configuration.

Configuration values are loaded from the following sources (in order of priority):

1. Configuration files (JSON, YAML, TOML)
2. OS environment variables (take precedence)
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

- Supported value types: `string`, `bool`, numeric types (converted to `int64` or `float64` when stored)
- Values of unsupported types are ignored
- An environment variable reference can be used as a value: `${ENV_VARIABLE:DEFAULT_VALUE}`

The value type is determined either by the environment variable or by the default value.
The default value may be omitted (in this case the format is `${ENV_VARIABLE}`).

If a parameter's value is a string, variable substitution inside the string is supported:
```json
{
    "url": "${SRV_PROTOCOL:https}://${SRV_NAME}:${SRV_PORT:443}/api"
}
```

### `.env` File Support

The library automatically loads environment variables from a `.env` file in the application's working directory (if it exists).

Format:
```
# Comment lines start with #
DB_HOST=localhost
DB_PORT=5432
DEBUG=true
```

- OS environment variables take precedence over `.env` file values
- The `.env` file is optional and silently ignored if not found
- The `KEY=VALUE` format with optional quotes is supported

### JSON and YAML

- The root element must be an object
- Using an array as the root element is not supported

---

## Usage

### Configuration Options

The `Options` struct is used to configure the configuration manager:

```go
type Options struct {
    Path          string          // Path to the directory containing the configuration file
    Source        string          // Path to the configuration (when adding custom readers, the path may refer not only to a file)
    Filename      string          // Name of the configuration file
    Format        string          // Data format of the configuration (built-in formats: json, yaml or toml)
    Reader        Reader          // Custom configuration reader (see the Reader interface)
    Parser        Parser          // Custom configuration parser (see the Parser interface)
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

- `-config.path=/opt/app/config` — sets the path to the directory with the configuration file (must not end with a "/" separator)
- `-config.filename=config.yml` — sets the name of the configuration file
- `-config.format=yaml` — sets the format of the configuration content
- `-config.source=/opt/app/config/config.yml` — sets the full path to the configuration file (when this flag is specified, the `config.path` and `config.filename` flags are not used) or another location of the configuration data, for example, the URL of an API method that returns the configuration data (in this case, a custom reader implementation is required)

---

### Configuration Parsers

Three parsers are available (registered by default):

- JSON parser (for `Format = "json"`)
- YAML parser (for `Format = "yaml"`)
- TOML parser (for `Format = "toml"`)

You can create a custom parser (implement the `Parser` interface) and register it using the `RegisterParser(format string, parser Parser)` method.

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

Creates a new configuration manager with the specified options. Pass `nil` to use the default options.
However, before using the configuration, you must call the `Apply()` method.

#### Methods

| Method | Description |
|------|----------|
| `Init(options *Options)` | Overrides the configuration options |
| `Apply(ctx context.Context)` | Loads and applies the configuration |
| `Add(key string, valueType ValueDataType) *Parameter` | Adds a parameter without setting a value |
| `Set(key string, value any)` | Sets a parameter value (or adds it if it does not exist) |
| `Get(key string) *Parameter` | Gets a configuration parameter |
| `GetAll() []*Parameter` | Gets all parameters |
| `Lookup(filter string) []*Parameter` | Gets an array of parameters by prefix |
| `Bind(object any)` | Binds a struct to the configuration |
| `BindFrom(object any, prefix string)` | Binds a struct to the configuration using relative paths to configuration parameters in the `config` tag. |
| | The prefix sets the starting point from which the configuration is read for the struct |
| `GetArrayValues(key string) map[string]any` | Gets the values of an array parameter (as "key-value") |
| `GetIntArray(key string) []int64` | Gets the values for an array parameter with integer values |
| `GetFloatArray(key string) []float64` | Gets the values for an array parameter with floating-point values |
| `GetStringArray(key string) []string` | Gets the string values for an array parameter |

---

### Parameter

The type of a configuration parameter.

```go
type Parameter struct {
    // Contains configuration parameter data with thread-safe access
}
```

#### Methods

| Method | Description |
|------|----------|
| `Name() string` | Gets the parameter name |
| `Value() any` | Gets the raw value of the parameter |
| `Type() reflect.Kind` | Gets the generic type of the parameter value |
| `Int() int64` | Gets the parameter value as int64 |
| `Float() float64` | Gets the parameter value as float64 |
| `Bool() bool` | Gets the parameter value as bool |
| `String() string` | Gets the parameter value as string |
| `SetValue(value any)` | Sets the parameter value with conversion to valid value types (int64, float64, bool, string) |

---

### StructuredConfiguration

A generic type for working with typed configuration structs.

```go
type StructuredConfiguration struct {
    C *configuration // Configuration
}
```

#### Functions

| Function | Description |
|--------|----------|
| `NewStructuredConfiguration[T any](options *Options) *T` | Creates a typed structured configuration (the passed type must embed the `StructuredConfiguration` struct) |

#### Methods

| Method | Description |
|------|----------|
| `Apply(ctx context.Context)` | Loads and applies the configuration |

---

### Reader

An interface for reading configuration data from a source.

```go
type Reader interface {
    Read(ctx context.Context, source string) ([]byte, error)
}
```

Implement this interface to create custom readers that can read the configuration from any source (files, HTTP, databases, etc.).

---

### Parser

An interface for parsing configuration data.

```go
type Parser interface {
    Parse(ctx context.Context, data []byte) (map[string]any, error)
}
```

Implement this interface to create custom parsers for different data formats. Register custom parsers via `RegisterParser(format string, parser Parser)`.

---

## Binding Structs to Configuration

Configuration parameter values can be bound to exported struct fields using the `config` struct tag.

Rules:

- Fields must be exported
- The field type must match the configuration value type
- If a parameter is missing from the file, an environment variable is used (if set)

There are two ways to bind configuration to structs:

### Method 1: Using the Bind Method

Create a struct with `config` tags, then use the `Bind` and `Apply` methods:

```go
package main

import (
    "context"

    "github.com/cvartan/goconfig"
)

type AppConfig struct {
    Name  string `config:"app.name"`
    Debug bool   `config:"app.debug"`
}

func main() {
    var cfg AppConfig

    config := goconfig.NewConfiguration(nil) // Use default options

    config.Bind(&cfg) // Populate the struct fields
    config.Apply(context.Background()) // Read the configuration from the file and apply the parameter values to the attributes of the bound structs
}
```

If the path to a parameter specified in the value of the `config` tag is not relative to the configuration root (for example, the configuration has a separate section used for a struct, which may vary from application to application), then in this case it is recommended to use the `BindFrom` method, which allows you to set a starting prefix — the point from which the configuration should be read for the struct.

### Method 2: Using StructuredConfiguration

Create a struct that embeds `goconfig.StructuredConfiguration`, then initialize it with `NewStructuredConfiguration`:

```go
package main

import (
    "context"

    "github.com/cvartan/goconfig"
)

type AppConfig struct {
    goconfig.StructuredConfiguration

    Name  string `config:"app.name"`
    Debug bool   `config:"app.debug"`
}

func main() {
    // Initialize the structured configuration with options (or nil for the default values)
    cfg := goconfig.NewStructuredConfiguration[AppConfig](nil)

    cfg.Apply(context.Background()) // Read the configuration from the file and apply it to the struct fields

    // Access the configuration values directly through the struct fields
    // fmt.Println(cfg.Name)
    // fmt.Println(cfg.Debug)
}
```

The `StructuredConfiguration` approach provides:

- Type-safe access to the configuration through the struct fields
- Direct access to the configuration methods via the embedded `StructuredConfiguration` (e.g., `cfg.Apply()`)
- Cleaner code organization by combining the configuration struct and behavior

---

### Struct Usage Limitations

- Struct fields that are slices or arrays are supported only if the array elements are simple values (not structs)

---

## License

MIT
