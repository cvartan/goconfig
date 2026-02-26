package tomlparser

import (
	"github.com/cvartan/goconfig/mapconv"
	toml "github.com/pelletier/go-toml/v2"
)

type TomlConfigurationParser struct{}

func (r *TomlConfigurationParser) Parse(data []byte) (props map[string]any, err error) {

	var propertySource map[string]any
	err = toml.Unmarshal(data, &propertySource)
	if err != nil {
		return
	}

	props = mapconv.ParseMapToPropertyMap(propertySource)

	return
}
