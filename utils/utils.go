package utils

import (
	"os"
	"reflect"
	"strconv"
)

func GetWD() string {
	if wd, err := os.Getwd(); err == nil {
		return wd
	}

	panic("[config:getWd:01] Can't get working directory")
}

// Type simplification - converting types to the set of types used for configuration parameters in this package
func TypeSimplified(source reflect.Kind) reflect.Kind {
	switch source {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		{
			return reflect.Int64
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		{
			return reflect.Uint64
		}
	case reflect.Float64, reflect.Float32:
		{
			return reflect.Float64
		}
	default:
		{
			return source
		}
	}
}

func StringValueToTypedValue(value string) any {
	if value == "" {
		return value
	}

	//Check boolean
	if bv, err := strconv.ParseBool(value); err == nil {
		return bv
	}

	//Check int64
	if iv, err := strconv.Atoi(value); err == nil {
		return iv
	}

	//Check float64
	if fv, err := strconv.ParseFloat(value, 64); err == nil {
		return fv
	}

	return value
}
