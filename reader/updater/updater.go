package updater

import (
	"os"
	"strconv"
	"strings"
)

func ParseEnvPlaveholders(props map[string]any) {
	for k, v := range props {
		switch str := v.(type) {
		case string:
			{
				if strings.HasPrefix(str, "${") {
					buf := strings.TrimFunc(
						str,
						func(r rune) bool {
							switch string(r) {
							case "$", "{", "}":
								{
									return true
								}
							default:
								{
									return false
								}
							}
						},
					)

					var envValue, defValue string

					parts := strings.Split(buf, ":")
					if len(parts) > 0 {
						envValue = parts[0]
					}
					if len(parts) > 1 {
						defValue = parts[1]
					}

					paramValue := os.Getenv(envValue)
					if paramValue == "" {
						paramValue = defValue
					}

					if tv, err := strconv.ParseBool(paramValue); err == nil {
						props[k] = tv
						continue
					}

					if tv, err := strconv.Atoi(paramValue); err == nil {
						props[k] = tv
						continue
					}

					props[k] = paramValue
				}
			}
		}
	}
}

func ReplaceValueByEnv(props map[string]any) {
	for k, v := range props {
		envVar := strings.ToUpper(strings.ReplaceAll(k, ".", "_"))
		envVal := os.Getenv(envVar)
		if envVal != "" {
			switch v.(type) {
			case bool:
				{
					if val, err := strconv.ParseBool(envVal); err == nil {
						props[k] = val
					}
				}
			case int:
				{
					if val, err := strconv.Atoi(envVal); err != nil {
						props[k] = val
					}
				}
			default:
				{
					props[k] = envVal
				}
			}
		}
	}
}
