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

	if i, ok := checkStructuredConfiguration(holding); ok {
		conf := &configuration{}
		conf.init(options)
		conf.bind(holding)

		holdingValue.Field(i).Field(0).Set(reflect.ValueOf(conf))
		return holding
	}

	return nil
}

func checkStructuredConfiguration(obj any) (int, bool) {
	objType := reflect.ValueOf(obj).Elem().Type()
	for i := 0; i < objType.NumField(); i++ {
		f := objType.Field(i)
		if f.Anonymous && f.Type == nestedConfigType {
			return i, true
		}
	}
	return -1, false
}
