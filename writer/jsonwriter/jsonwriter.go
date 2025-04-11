package jsonwriter

import (
	"encoding/json"
	"os"

	"github.com/cvartan/goconfig/mapconv"
)

type JsonWriter struct {
}

func (w *JsonWriter) Write(filename string, props map[string]any) error {
	sm := mapconv.ParsePropertyMapToMap(props)

	data, err := json.MarshalIndent(sm, "", "\t")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0666)
}
