package goconfig

import "reflect"

type StructuredConfiguration struct {
	C *configuration // Configuration
}

func (s *StructuredConfiguration) Apply() {
	if s.C == nil {
		panic("Configuration is not defined. Before, use GenerateStructuredConfiguration function for create structured configuration.")
	}
	s.C.apply()
}

var nestedConfigType reflect.Type = reflect.ValueOf(StructuredConfiguration{}).Type()

func NewStructuredConfiguration[T any](options *Options) *T {
	holding := new(T)
	holdingValue := reflect.ValueOf(holding).Elem()
	holdingType := holdingValue.Type()

	for i := 0; i < holdingType.NumField(); i++ {
		f := holdingType.Field(i)
		if f.Anonymous && f.Type == nestedConfigType {
			conf := &configuration{}
			conf.init(options)
			conf.bind(holding)

			holdingValue.Field(i).Field(0).Set(reflect.ValueOf(conf))
			return holding
		}
	}

	return nil
}
