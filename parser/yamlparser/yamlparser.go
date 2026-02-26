package yamlparser

import (
	"github.com/cvartan/goconfig/mapconv"
	"gopkg.in/yaml.v3"
)

type YamlConfigurationParser struct{}

func (r *YamlConfigurationParser) Parse(data []byte) (props map[string]any, err error) {

	var propertySource map[string]any
	err = yaml.Unmarshal(data, &propertySource)

	if err != nil {
		return
	}

	props = mapconv.ParseMapToPropertyMap(propertySource)

	return
}
