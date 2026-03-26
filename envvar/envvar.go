package envvar

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/cvartan/goconfig/utils"
)

type EnvVariables struct {
	envFileVars map[string]string
}

func NewEnvVariables() *EnvVariables {
	e := &EnvVariables{
		envFileVars: make(map[string]string),
	}

	wd := utils.GetWD()
	envPath := wd + "/.env"

	file, err := os.Open(envPath)
	if err == nil {
		// .env file is optional, silently ignore if not found

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())

			// Skip empty lines and comments
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}

			// Parse KEY=VALUE format
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Remove quotes if present
			value = strings.Trim(value, `"'`)

			e.envFileVars[key] = value
		}
	}
	file.Close()

	return e
}

func (e *EnvVariables) GetValue(key string) string {
	// First check OS environment variable
	if val := os.Getenv(key); val != "" {
		return val
	}

	// Fallback to .env file value
	if e.envFileVars != nil {
		if val, ok := e.envFileVars[key]; ok {
			return val
		}
	}

	return ""
}

func (e *EnvVariables) GetValueForType(key string, targetType reflect.Kind) any {
	v := e.GetValue(key)
	if v == "" {
		return nil
	}
	switch utils.TypeSimplified(targetType) {
	case reflect.Int64, reflect.Uint64:
		{
			tv, err := strconv.Atoi(v)
			if err != nil {
				panic(fmt.Sprintf("Can't convert string value (%s) to int", v))
			}
			return tv
		}
	case reflect.Float64:
		{
			tv, err := strconv.ParseFloat(v, 64)
			if err != nil {
				panic(fmt.Sprintf("Can't convert string value (%s) to float", v))
			}
			return tv
		}
	case reflect.Bool:
		{
			tv, err := strconv.ParseBool(v)
			if err != nil {
				panic(fmt.Sprintf("Can't convert string value (%s) to boolean", v))
			}
			return tv
		}
	default:
		{
			return v
		}
	}
}
