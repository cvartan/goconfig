package yamlparser

import (
	"context"

	"github.com/cvartan/goconfig/mapconv"
	"github.com/cvartan/goconfig/types"
	"gopkg.in/yaml.v3"
)

type YamlConfigurationParser struct{}

func (r *YamlConfigurationParser) Parse(ctx context.Context, data []byte) (props map[string]any, err error) {

	var propertySource map[string]any
	err = yaml.Unmarshal(data, &propertySource)

	if err != nil {
		return nil, types.NewParseConfigurationDataError(err, "json")
	}

	props = mapconv.ParseMapToPropertyMap(propertySource)

	return
}
