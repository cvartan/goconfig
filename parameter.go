package goconfig

import (
	"fmt"
	"reflect"
	"strconv"
	"sync"

	"github.com/cvartan/goconfig/utils"
)

// Type for configuration parameter
type Parameter struct {
	mu                 sync.Mutex
	name               string
	value              any
	valueType          reflect.Kind
	bindedStructFields []*fieldListItem
}

// Get configuration parameter name
func (p *Parameter) Name() string {
	return p.name
}

// Get configuration parameter value
func (p *Parameter) Value() any {
	return p.value
}

// Get configuration parameter value type
func (p *Parameter) Type() reflect.Kind {
	return p.valueType
}

// Get an integer parameter value
func (p *Parameter) Int() (value int64) {
	switch p.Type() {
	case reflect.Int64:
		{
			return p.value.(int64)
		}
	default:
		{
			panic("[parameter.int.01.] value of the parameter is not int64")
		}
	}
}

// Get a float64 parameter value
func (p *Parameter) Float() (value float64) {
	switch p.Type() {
	case reflect.Int64:
		{
			return float64(p.value.(int64))
		}
	case reflect.Float64:
		{
			return p.value.(float64)
		}
	default:
		{
			panic("[parameter.float.01.] value of the parameter is not int64 or float64")
		}
	}
}

// Get a boolean parameter value
func (p *Parameter) Bool() (value bool) {
	switch p.Type() {
	case reflect.Bool:
		{
			return p.value.(bool)
		}
	default:
		{
			panic("[parameter.bool.01.] value of the parameter is not boolean")
		}
	}
}

// Get a string parameter value
func (p *Parameter) String() string {

	switch p.Type() {
	case reflect.Int64:
		{
			i := p.value.(int64)
			return strconv.FormatInt(i, 10)
		}
	case reflect.Float64:
		{
			f := p.value.(float64)
			return strconv.FormatFloat(f, 'g', -1, 64)
		}
	case reflect.Bool:
		{
			b := p.value.(bool)
			return strconv.FormatBool(b)
		}
	case reflect.String:
		{
			return p.value.(string)
		}
	}

	return fmt.Sprintf("%v", p.value)
}

func (p *Parameter) bindStructField(field *fieldListItem) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.bindedStructFields = append(p.bindedStructFields, field)
	field.setValue(p)
}

func (p *Parameter) SetValue(value any) {
	if value == nil {
		// Pass nil value
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	switch t := utils.TypeSimplified(reflect.TypeOf(value).Kind()); t {
	case reflect.Int64, reflect.Uint64, reflect.Float64, reflect.Bool, reflect.String:
		{
			tp := t
			if tp == reflect.Uint64 {
				tp = reflect.Int64
			}

			if p.Type() != tp {
				panic(fmt.Sprintf("[config:set:03] incompatible data types between property %s and value", p.name))
			}

			p.value = value

		}
	default:
		{
			panic(fmt.Sprintf("[config:set:04] unsupported value type: %v", t))
		}
	}

	// Set value for linked structure fields
	for _, f := range p.bindedStructFields {
		f.setValue(p)
	}
}
