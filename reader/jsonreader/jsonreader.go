package jsonreader

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"

	"github.com/cvartan/goconfig/mapconv"
)

type JsonConfigurationReader struct {
}

func (r *JsonConfigurationReader) Read(filename string) (props map[string]any, err error) {
	var fs []byte
	fs, err = os.ReadFile(filename)
	if err != nil {
		return
	}

	var propertySource map[string]any
	err = json.Unmarshal(fs, &propertySource)
	if err != nil {
		return
	}

	props = mapconv.ParseMapToPropertyMap(propertySource)

	// Устанавливаем значения свойств записанные как ссылка на переменную окружения
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
	// Заменяем значения свойств если есть переменная окружения, в соответствующем формате, в которой определено другое значение (например, для переопределения свойств из конфигурационного файла при запуске приложения в докере)
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

	return
}
