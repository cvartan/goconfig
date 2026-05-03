package tomlparser

import (
	"github.com/cvartan/goconfig/mapconv"
	"github.com/cvartan/goconfig/types"
	toml "github.com/pelletier/go-toml/v2"
)

type TomlConfigurationParser struct{}

func (r *TomlConfigurationParser) Parse(data []byte) (props map[string]any, err error) {

	var propertySource map[string]any
	err = toml.Unmarshal(data, &propertySource)
	if err != nil {
		return nil, types.NewParseConfigurationDataError(err, "json")
	}

	props = mapconv.ParseMapToPropertyMap(propertySource)

	return
}
