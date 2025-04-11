package yamlwriter

import (
	"os"

	"github.com/cvartan/goconfig/mapconv"
	"gopkg.in/yaml.v3"
)

type YamlConfigurationWriter struct {
}

func (w *YamlConfigurationWriter) Write(filename string, props map[string]any) error {
	sm := mapconv.ParsePropertyMapToMap(props)

	data, err := yaml.Marshal(sm)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0666)
}
