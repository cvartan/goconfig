package envvar

import (
	"bufio"
	"os"
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

func (e *EnvVariables) GetEnvValue(key string) string {
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
