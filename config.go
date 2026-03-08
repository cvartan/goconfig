package goconfig

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/cvartan/goconfig/parser/jsonparser"
	"github.com/cvartan/goconfig/parser/tomlparser"
	"github.com/cvartan/goconfig/parser/yamlparser"
	"github.com/cvartan/goconfig/reader/filereader"
)

const initMapLen int = 8

type Reader interface {
	Read(source string) ([]byte, error)
}

type Parser interface {
	Parse(data []byte) (map[string]any, error)
}

var parsers map[string]Parser

func RegisterParser(format string, parser Parser) {
	if format == "" {
		panic("[config:reg:01] format must be defined")
	}

	if parser == nil {
		panic("[config:reg:02] parser must be defined")
	}

	parsers[format] = parser
}

type Options struct {
	Path     string
	Source   string
	Filename string
	Format   string
	Reader   Reader
	Parser   Parser
}

// Main type for working with application configuration.
// Contains methods for accessing configuration parameters
type Configuration struct {
	mu                 sync.Mutex
	properties         map[string]*Parameter
	bindedStructFields []*fieldListItem
	reader             Reader
	parser             Parser
	source             string
}

type fieldListItem struct {
	parentStruct reflect.Value       // Structure the field belongs to
	field        reflect.StructField // Field
	fieldType    reflect.Kind        // Field data type
	configVar    string              // Configuration parameter associated with this field
}

var (
	defaultPath       string = *flag.String("config.path", getWD(), "path to the configuration file")
	defaultFormat     string = *flag.String("config.format", "json", "format of the configuration file")
	defaultFileName   string = *flag.String("config.filename", "config."+defaultFormat, "name of the configuration file")
	defaultFileSource string = *flag.String("config.source", "", "path to configuration sources")
)

func getWD() string {
	if wd, err := os.Getwd(); err == nil {
		return wd
	}

	panic("[config:getWd:01] Can't get working directory")
}

// Type for configuration parameter
type Parameter struct {
	name      string
	value     any
	valueType reflect.Kind
}

type ValueDataType int

const (
	String ValueDataType = iota
	Int
	Float
	Bool
)

func (t ValueDataType) Kind() reflect.Kind {
	switch t {
	case 0:
		{
			return reflect.String
		}
	case 1:
		{
			return reflect.Int64
		}
	case 2:
		{
			return reflect.Float64
		}
	case 3:
		{
			return reflect.Bool
		}
	}
	panic("[config:getType:001] unsupported data type")
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

func NewConfiguration(options *Options) *Configuration {

	opt := options

	if options == nil {
		opt = &Options{}
	}

	if opt.Format == "" {
		opt.Format = defaultFormat
	}

	if opt.Path == "" {
		opt.Path = defaultPath
	}

	if opt.Filename == "" {
		opt.Filename = defaultFileName
	}

	if opt.Source == "" {
		if defaultFileSource == "" {
			opt.Source = fmt.Sprintf("%s%s%s", opt.Path, string(os.PathSeparator), opt.Filename)
		} else {
			opt.Source = defaultFileSource
		}
	}

	if opt.Reader == nil {
		opt.Reader = &filereader.FileReader{}
	}

	if opt.Parser == nil {
		parser, ok := parsers[strings.ToLower(opt.Format)]
		if !ok {
			panic("[config:new:01] reader for this format is not defined")
		}
		opt.Parser = parser
	}

	cm := &Configuration{
		properties: make(map[string]*Parameter, initMapLen),
		reader:     opt.Reader,
		parser:     opt.Parser,
		source:     opt.Source,
	}

	return cm
}

// Add or modify parameter value
func (cm *Configuration) setParameterValue(key string, value any) {
	if value == nil {
		panic(fmt.Sprintf("[config:set:03] value for key=%s must be defined", key))
	}

	cm.mu.Lock()
	defer cm.mu.Unlock()
	switch t := typeSimplified(reflect.TypeOf(value).Kind()); t {
	case reflect.Int64, reflect.Uint64, reflect.Float64, reflect.Bool, reflect.String:
		{
			tp := t
			if tp == reflect.Uint64 {
				tp = reflect.Int64
			}

			// Check existing parameters
			if p, ok := cm.properties[key]; ok {

				if p.Type() != tp {
					panic(fmt.Sprintf("[config:set:03] incompatible data types between property %s and value", key))
				}
			}

			cv := &Parameter{
				name:      key,
				value:     value,
				valueType: tp,
			}

			replaceByEnvValues(cv)

			cm.properties[key] = cv
		}
	default:
		{
			panic(fmt.Sprintf("[config:set:04] unsupported value type: %v", t))
		}
	}
}

func (cm *Configuration) Add(key string, valueType ValueDataType) {
	t := valueType.Kind()

	if p, ok := cm.properties[key]; ok {
		if p.valueType != t {
			panic("[config:defParameter:001] parameter exists with incompatible data types")
		}
		return
	}

	p := &Parameter{
		name:      key,
		valueType: t,
		value:     nil,
	}

	replaceByEnvValues(p)

	cm.properties[key] = p
}

func (cm *Configuration) Set(key string, value any) {
	if key == "" {
		panic("[config:set:01] key is empty")
	}

	if value == nil {
		panic("[config:set:02] value is empty")
	}

	cm.setParameterValue(key, value)

}

// Load and apply the configuration
func (cm *Configuration) Apply() {

	if cm.reader == nil {
		panic("[config:read:01] configuration reader is not defined")
	}

	if cm.parser == nil {
		panic("[config.read:04] configuration parser is not defined")
	}

	fs, err := cm.reader.Read(cm.source)
	if err != nil {
		panic(fmt.Sprintf("[config:read:02] can't read configuration by error: %s", err))
	}

	props, err := cm.parser.Parse(fs)
	if err != nil {
		panic(fmt.Sprintf("[config:read:05] can't parse configuration by error: %s", err))
	}

	// Set property values written as references to environment variables
	parseEnvPlaceholders(props)

	// Perform correction of numeric values - converting them to int64 or float64 types
	// For the case when all numeric values are converted to one type when reading from the configuration source (e.g., to float64 during JSON deserialization), the correct type is determined by value during reading (reader's Read method).
	correctNumberValues(props)

	// Fill cm.properties with values from received properties
	for k, v := range props {
		cm.setParameterValue(k, v)
	}

	cm.fillStructFields()
}

// Replaces value if the configuration parameter value was specified in format ${env_value:default_value}
func parseEnvPlaceholders(props map[string]any) {
	for k, v := range props {
		switch str := v.(type) {
		case string:
			{
				pc := strings.Count(str, "${")
				if pc > 0 {
					if pc == 1 && strings.HasPrefix(str, "${") && strings.HasSuffix(str, "}") {
						props[k] = parseValuePlaceholder(str)
						continue
					}

					values := extractPlaceholders(str)
					for _, v := range values {
						value := parseValuePlaceholder(v)
						str = strings.Replace(str, "${"+v+"}", fmt.Sprintf("%v", value), 1)
					}

					props[k] = str
				}
			}
		}
	}
}

func extractPlaceholders(str string) (result []string) {
	result = make([]string, 0, strings.Count(str, "${"))

	vs := strings.Split(str, "${")
	for i := 1; i < len(vs); i++ {
		var buf strings.Builder
		for _, a := range vs[i] {
			if a == '}' {
				result = append(result, buf.String())
				break
			}
			buf.WriteRune(a)
		}
	}

	return
}

func parseValuePlaceholder(str string) any {
	buf := strings.TrimFunc(
		str,
		func(r rune) bool {
			switch string(r) {
			case "$", "{", "}":
				{
					return true
				}
			default:
				{
					return false
				}
			}
		},
	)

	var envValue, defValue string

	parts := strings.Split(buf, ":")
	if len(parts) > 0 {
		envValue = parts[0]
	}
	if len(parts) > 1 {
		defValue = parts[1]
	}

	paramValue := os.Getenv(strings.ToUpper(envValue))
	if paramValue == "" {
		paramValue = defValue
	}

	if tv, err := strconv.ParseBool(paramValue); err == nil {
		return tv
	}

	if tv, err := strconv.Atoi(paramValue); err == nil {
		return tv
	}

	return paramValue
}

func correctNumberValues(props map[string]any) {
	for k, v := range props {
		switch n := v.(type) {
		case int:
			{
				props[k] = int64(n)
			}
		case int8:
			{
				props[k] = int64(n)
			}
		case int16:
			{
				props[k] = int64(n)
			}
		case int32:
			{
				props[k] = int64(n)
			}
		case int64:
			{
				props[k] = n
			}
		case uint:
			{
				props[k] = int64(n)
			}
		case uint8:
			{
				props[k] = int64(n)
			}
		case uint16:
			{
				props[k] = int64(n)
			}
		case uint32:
			{
				props[k] = int64(n)
			}
		case uint64:
			{
				props[k] = int64(n)
			}
		case float64:
			{
				props[k] = n
			}
		case float32:
			{
				props[k] = float64(n)
			}

		}
	}
}

// Replace property values if there is an environment variable in the corresponding format that defines a different value (e.g., for overriding properties from the configuration file when running the application in Docker)
func replaceByEnvValues(value *Parameter) {

	envVar := strings.ReplaceAll(value.name, ".", "_")
	envVal := os.Getenv(strings.ToUpper(envVar))
	if envVal != "" {
		switch value.Type() {
		case reflect.Int64:
			{
				if val, err := strconv.ParseInt(envVal, 10, 64); err == nil {
					value.value = val
					value.valueType = reflect.Int64
					return
				}

			}
		case reflect.Bool:
			{
				if val, err := strconv.ParseBool(envVal); err == nil {
					value.value = val
					value.valueType = reflect.Bool
					return
				}
			}
		case reflect.Float64:
			{
				if val, err := strconv.ParseFloat(envVal, 64); err == nil {
					value.value = val
					value.valueType = reflect.Float64
					return
				}
			}
		default:
			{
				value.value = envVal
				value.valueType = reflect.String
				return
			}
		}
	}

}

// Get a parameter value
func (cm *Configuration) Get(key string) *Parameter {
	return cm.properties[key]
}

// Get all parameters
func (cm *Configuration) GetAll() (parameters []*Parameter) {
	parameters = make([]*Parameter, 0, len(cm.properties))
	for _, v := range cm.properties {
		parameters = append(parameters, v)
	}
	return
}

// Get parameters by prefix
func (cm *Configuration) Lookup(filter string) (parameters []*Parameter) {
	parameters = make([]*Parameter, 0, len(cm.properties))
	for _, v := range cm.properties {
		if strings.HasPrefix(v.Name(), strings.ToLower(filter)) {
			parameters = append(parameters, v)
		}
	}
	return
}

func checkEndingDot(param string) string {
	if param[len(param)-1] != '.' {
		return param + "."
	}

	return param
}

// Get values for an array parameter
func (cm *Configuration) GetArrayValues(key string) (values []any) {
	parameters := cm.Lookup(checkEndingDot(key))
	values = make([]any, 0, len(parameters))
	for _, v := range parameters {
		values = append(values, v.value)
	}
	return
}

// Get integer values for an array parameter
func (cm *Configuration) GetIntArray(key string) (values []int64) {
	parameters := cm.Lookup(checkEndingDot(key))
	values = make([]int64, 0, len(parameters))

	for _, p := range parameters {
		values = append(values, p.Int())
	}
	return
}

// Get float values for an array parameter
func (cm *Configuration) GetFloatArray(key string) (values []float64) {
	parameters := cm.Lookup(checkEndingDot(key))
	values = make([]float64, 0, len(parameters))

	for _, p := range parameters {
		values = append(values, p.Float())
	}
	return
}

// Get string values for an array parameter
func (cm *Configuration) GetStringArray(key string) (values []string) {
	parameters := cm.Lookup(checkEndingDot(key))
	values = make([]string, 0, len(parameters))

	for _, p := range parameters {
		values = append(values, p.String())
	}
	return
}

// Bind struct (marked with the config tag) to configuration
// Rules:
// - Fields must be exported
// - Field types must match the configuration value types
// - If a parameter is missing in the file, the environment variable is used (if defined)
//
// Example:
//
//	type Config struct {
//	    Attr1 string `config:"app.attr1"`
//	    ...
//	}
//
// Parameter object must be pointer to struct
func (cm *Configuration) Bind(object any) {
	if object == nil {
		panic("[config:bind:01] binded object must be defined")
	}
	c := reflect.ValueOf(object).Elem()
	if c.Type().Kind() != reflect.Struct {
		panic("[config:bind:02] binded object is not struct")
	}

	// Collect all fields, including those in nested structures
	for _, f := range reflect.VisibleFields(c.Type()) {
		switch f.Type.Kind() {
		case reflect.Struct:
			{
				cm.collectStructFields(c.FieldByName(f.Name), f)
			}
		case reflect.Array, reflect.Slice:
			{
				// TODO: apparently someday we need to figure out how to work with arrays in struct attributes
				continue
			}
		case reflect.String, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
			{
				cm.bindStructField(
					&fieldListItem{
						parentStruct: c,
						field:        f,
						fieldType:    typeSimplified(f.Type.Kind()),
					},
				)
			}
		}
	}
}

// Getting fields of nested structure
func (cm *Configuration) collectStructFields(structValue reflect.Value, structField reflect.StructField) {
	fields := reflect.VisibleFields(structField.Type)
	for _, f := range fields {
		switch f.Type.Kind() {
		case reflect.Struct:
			{
				cm.collectStructFields(structValue.FieldByName(f.Name), f)
			}
		case reflect.Array, reflect.Slice:
			{
				// TODO: apparently someday we need to figure out how to work with arrays in struct attributes
				continue
			}
		case reflect.String, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
			{
				cm.bindStructField(&fieldListItem{
					parentStruct: structValue,
					field:        f,
					fieldType:    typeSimplified(f.Type.Kind()),
				})
			}
		}
	}
}

// Check struct field for config tag and add field to list of bound fields for subsequent filling after configuration reading
func (cm *Configuration) bindStructField(f *fieldListItem) {
	tagValue := f.field.Tag.Get("config")
	if tagValue != "" {
		cm.mu.Lock()
		cm.bindedStructFields = append(cm.bindedStructFields, &fieldListItem{
			parentStruct: f.parentStruct,
			field:        f.field,
			fieldType:    f.fieldType,
			configVar:    tagValue,
		})
		cm.mu.Unlock()

		// Check that existing property has same data type as struct attribute
		if p, ok := cm.properties[tagValue]; ok {
			tp := f.fieldType
			if tp == reflect.Uint64 {
				tp = reflect.Int64
			}

			if p.Type() != tp {
				panic(fmt.Sprintf("[config:bindStructField:01] uncompatible data type between parameter %s and field %s", tagValue, f.field.Name))
			}

			return
		}

		// Add new property
		cm.mu.Lock()
		switch f.fieldType {
		case reflect.Int64, reflect.Uint64:
			{
				cv := &Parameter{
					name:      tagValue,
					value:     f.parentStruct.FieldByName(f.field.Name).Int(),
					valueType: reflect.Int64,
				}

				replaceByEnvValues(cv)

				cm.properties[tagValue] = cv
			}
		case reflect.Float64:
			{
				cv := &Parameter{
					name:      tagValue,
					value:     f.parentStruct.FieldByName(f.field.Name).Float(),
					valueType: reflect.Float64,
				}

				replaceByEnvValues(cv)

				cm.properties[tagValue] = cv
			}
		case reflect.Bool:
			{
				cv := &Parameter{
					name:      tagValue,
					value:     f.parentStruct.FieldByName(f.field.Name).Bool(),
					valueType: reflect.Bool,
				}

				replaceByEnvValues(cv)

				cm.properties[tagValue] = cv
			}
		case reflect.String:
			{
				cv := &Parameter{
					name:      tagValue,
					value:     f.parentStruct.FieldByName(f.field.Name).String(),
					valueType: reflect.String,
				}

				replaceByEnvValues(cv)

				cm.properties[tagValue] = cv
			}
		}
		cm.mu.Unlock()
	}
}

// Filling fields of bound structures with configuration parameter values
func (cm *Configuration) fillStructFields() {
	// Iterate through configuration struct fields and set values from received property list
	for _, f := range cm.bindedStructFields {
		tagValue := f.configVar
		if propValue, ok := cm.properties[tagValue]; ok {
			switch f.fieldType {
			case reflect.Float64:
				{
					if propValue.Type() != reflect.Float64 {
						panic(fmt.Sprintf("[config:struct:01] incorrect type for field %s (must be float)", f.field.Name))
					}
					floatval, ok := propValue.Value().(float64)
					if !ok {
						panic(fmt.Sprintf("[config:struct:02] incorrect type conversion for float field %s", f.field.Name))
					}
					floatf := f.parentStruct.FieldByName(f.field.Name)
					if floatf.OverflowFloat(floatval) {
						panic(fmt.Sprintf("[config:struct:03] can't set float64 value %f for float field %s", floatval, f.field.Name))
					}
					floatf.SetFloat(floatval)
				}
			case reflect.Int64:
				{
					if propValue.Type() != reflect.Int64 {
						panic(fmt.Sprintf("[config:struct:04] incorrect type for field %s (must be integer)", f.field.Name))
					}

					intval := propValue.Value().(int64)

					intf := f.parentStruct.FieldByName(f.field.Name)
					if intf.OverflowInt(intval) {
						panic(fmt.Sprintf("[config:struct:05] can't set int64 value %d for integer field %s", intval, f.field.Name))
					}
					intf.SetInt(intval)
				}
			case reflect.Uint64:
				{
					if propValue.Type() != reflect.Int64 {
						panic(fmt.Sprintf("[config:struct:06] incorrect type for field %s (must be integer)", f.field.Name))
					}

					intval := uint64(propValue.Value().(int64))

					intf := f.parentStruct.FieldByName(f.field.Name)
					if intf.OverflowUint(intval) {
						panic(fmt.Sprintf("[config:struct:07] can't set uint64 value %d for unsigned integer field %s", intval, f.field.Name))
					}
					intf.SetUint(intval)
				}
			case reflect.Bool:
				{
					if propValue.Type() != reflect.Bool {
						panic(fmt.Sprintf("[config:struct:08] incorrect type for field %s (must be boolean)", f.field.Name))
					}
					boolval, ok := propValue.Value().(bool)
					if !ok {
						panic(fmt.Sprintf("[config:struct:09] incorrect type conversion for boolean field %s", f.field.Name))
					}
					f.parentStruct.FieldByName(f.field.Name).SetBool(boolval)
				}
			case reflect.String:
				{
					if propValue.Type() != reflect.String {
						panic(fmt.Sprintf("[config:struct:10] incorrect type for field %s (must be string)", f.field.Name))
					}
					strval, ok := propValue.Value().(string)
					if !ok {
						panic(fmt.Sprintf("[config:struct:11] incorrect type conversion for string field %s", f.field.Name))
					}
					f.parentStruct.FieldByName(f.field.Name).SetString(strval)
					f.fieldType = reflect.String
				}
			}
		}
	}
}

// Type simplification - converting types to the set of types used for configuration parameters in this package
func typeSimplified(source reflect.Kind) reflect.Kind {
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

func init() {
	parsers = make(map[string]Parser, 1)
	RegisterParser("json", &jsonparser.JsonConfigurationParser{})
	RegisterParser("yaml", &yamlparser.YamlConfigurationParser{})
	RegisterParser("toml", &tomlparser.TomlConfigurationParser{})
}
