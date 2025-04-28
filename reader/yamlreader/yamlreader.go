package yamlreader

import (
	"os"

	"github.com/cvartan/goconfig/mapconv"
	"github.com/cvartan/goconfig/reader/updater"
	"gopkg.in/yaml.v3"
)

type YamlConfigurationReader struct{}

func (r *YamlConfigurationReader) Read(filename string) (props map[string]any, err error) {
	var fs []byte
	fs, err = os.ReadFile(filename)
	if err != nil {
		return
	}

	var propertySource map[string]any
	err = yaml.Unmarshal(fs, &propertySource)

	if err != nil {
		return
	}

	props = mapconv.ParseMapToPropertyMap(propertySource)

	// Устанавливаем значения свойств записанные как ссылка на переменную окружения
	updater.ParseEnvPlaceholders(props)
	// Заменяем значения свойств если есть переменная окружения, в соответствующем формате, в которой определено другое значение (например, для переопределения свойств из конфигурационного файла при запуске приложения в докере)
	updater.ReplaceValueByEnv(props)

	return
}
