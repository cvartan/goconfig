package utils

type Ints interface {
	int | int8 | int16 | int32 | int64 | uint8 | uint16 | uint32 | uint64
}

func typedIntConv[T Ints](v T) int64 {
	return int64(v)
}

func AnyToInt64(v any) int64 {
	switch t := v.(type) {
	case int:
		{
			return typedIntConv(t)
		}
	case int8:
		{
			return typedIntConv(t)
		}
	case int16:
		{
			return typedIntConv(t)
		}
	case int32:
		{
			return typedIntConv(t)
		}
	case uint8:
		{
			return typedIntConv(t)
		}
	case uint16:
		{
			return typedIntConv(t)
		}
	case uint32:
		{
			return typedIntConv(t)
		}
	case int64:
		{
			return t
		}
	case uint64:
		{
			return typedIntConv(t)
		}

	}

	return int64(0)
}

type Floats interface {
	float32 | float64
}

func typedFloatConv[T Floats](v T) float64 {
	return float64(v)
}

func AnyToFloat64(v any) float64 {
	switch t := v.(type) {
	case float32:
		{
			return typedFloatConv(t)
		}
	case float64:
		{
			return t
		}
	}
	return float64(0)
}
