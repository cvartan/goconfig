package jsonparser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cvartan/goconfig/mapconv"
	"github.com/cvartan/goconfig/types"
)

type JsonConfigurationParser struct {
}

func (r *JsonConfigurationParser) Parse(data []byte) (props map[string]any, err error) {

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	var propertySource map[string]any
	err = decoder.Decode(&propertySource)
	if err != nil {
		return nil, types.NewParseConfigurationDataError(err, "json")
	}

	props = mapconv.ParseMapToPropertyMap(propertySource)

	for k, v := range props {
		switch t := v.(type) {
		case json.Number:
			{
				if !strings.Contains(string(t), ".") {
					val, err := t.Int64()
					if err != nil {
						panic(fmt.Sprintf("[config:jsonparser:01] can't convert value %v to Int64 (%v)", t, k))
					}
					props[k] = val
				} else {
					val, err := t.Float64()
					if err != nil {
						panic(fmt.Sprintf("[config:jsonparser:02] can't convert value %v to Float64 (%v)", t, k))
					}
					props[k] = val
				}
			}
		}
	}

	return
}
