package updater

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func ParseEnvPlaceholders(props map[string]any) {
	for k, v := range props {
		switch str := v.(type) {
		case string:
			{
				pc := strings.Count(str, "${")
				if pc > 0 {
					if pc == 1 && strings.HasPrefix(str, "${") && strings.HasSuffix(str, "}") {
						props[k] = parseValuePlaceholder(str)
						continue
					}

					values := extractPlaceholders(str)
					for _, v := range values {
						value := parseValuePlaceholder(v)
						str = strings.Replace(str, "${"+v+"}", fmt.Sprintf("%v", value), 1)
					}

					props[k] = str
				}
			}
		}
	}
}

func extractPlaceholders(str string) (result []string) {
	result = make([]string, 0, strings.Count(str, "${"))

	vs := strings.Split(str, "${")
	for i := 1; i < len(vs); i++ {
		var buf strings.Builder
		for _, a := range vs[i] {
			if a == '}' {
				result = append(result, buf.String())
				break
			}
			buf.WriteRune(a)
		}
	}

	return
}

func parseValuePlaceholder(str string) any {
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
		return tv
	}

	if tv, err := strconv.Atoi(paramValue); err == nil {
		return tv
	}

	return paramValue
}

func ReplaceValueByEnv(props map[string]any) {
	for k, v := range props {
		envVar := strings.ReplaceAll(k, ".", "_")
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
