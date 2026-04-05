package goconfig

import (
	"context"
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/cvartan/goconfig/envvar"
	"github.com/cvartan/goconfig/parser/jsonparser"
	"github.com/cvartan/goconfig/parser/tomlparser"
	"github.com/cvartan/goconfig/parser/yamlparser"
	"github.com/cvartan/goconfig/reader/filereader"
	"github.com/cvartan/goconfig/utils"
)

const (
	initMapLen int = 8
	initArrayLen
)

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
	Path          string          // Path to catalog with the configuration file
	Source        string          // Path to the configuration (may refer to non-file sources with custom readers)
	Filename      string          // Filename of the configuration file
	Format        string          // Data format in the configuration (default supported formats: json, yaml or toml)
	Reader        Reader          // Custom configuration reader (see Reader interface)
	ReaderOptions any             // Custom reader configuration (see used reader documentation)
	ReaderTimeout int             // Reading timeout (in milliseconds)
	Parser        Parser          // Custom configuration parser (see Parser interface)
	ParserOptions any             // Custom parser configuration (see used parser documentation)
	ParserTimeout int             // Parisng timeout (in millseconds)
	Context       context.Context // Application context
}

// Main type for working with application configuration.
// Contains methods for accessing configuration parameters
type Configuration struct {
	c configuration
}

type configuration struct {
	mu         sync.Mutex
	properties map[string]*Parameter
	reader     Reader
	parser     Parser
	source     string
	envVars    *envvar.EnvVariables
}

type fieldListItem struct {
	parentStruct reflect.Value       // Structure the field belongs to
	field        reflect.StructField // Field
	fieldType    reflect.Kind        // Field data type
	configVar    string              // Configuration parameter associated with this field
}

func (f *fieldListItem) setValue(value *Parameter) {
	if value.value == nil {
		return
	}
	switch f.fieldType {
	case reflect.Float64:
		{
			if value.Type() != reflect.Float64 {
				panic(fmt.Sprintf("[config:struct:01] incorrect type for field %s (must be float)", f.field.Name))
			}
			floatval, ok := value.Value().(float64)
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
			if value.Type() != reflect.Int64 {
				panic(fmt.Sprintf("[config:struct:04] incorrect type for field %s (must be integer)", f.field.Name))
			}

			intval := value.Value().(int64)

			intf := f.parentStruct.FieldByName(f.field.Name)
			if intf.OverflowInt(intval) {
				panic(fmt.Sprintf("[config:struct:05] can't set int64 value %d for integer field %s", intval, f.field.Name))
			}
			intf.SetInt(intval)
		}
	case reflect.Uint64:
		{
			if value.Type() != reflect.Int64 {
				panic(fmt.Sprintf("[config:struct:06] incorrect type for field %s (must be integer)", f.field.Name))
			}

			intval := uint64(value.Value().(int64))

			intf := f.parentStruct.FieldByName(f.field.Name)
			if intf.OverflowUint(intval) {
				panic(fmt.Sprintf("[config:struct:07] can't set uint64 value %d for unsigned integer field %s", intval, f.field.Name))
			}
			intf.SetUint(intval)
		}
	case reflect.Bool:
		{
			if value.Type() != reflect.Bool {
				panic(fmt.Sprintf("[config:struct:08] incorrect type for field %s (must be boolean)", f.field.Name))
			}
			boolval, ok := value.Value().(bool)
			if !ok {
				panic(fmt.Sprintf("[config:struct:09] incorrect type conversion for boolean field %s", f.field.Name))
			}
			f.parentStruct.FieldByName(f.field.Name).SetBool(boolval)
		}
	case reflect.String:
		{
			if value.Type() != reflect.String {
				panic(fmt.Sprintf("[config:struct:10] incorrect type for field %s (must be string)", f.field.Name))
			}
			strval, ok := value.Value().(string)
			if !ok {
				panic(fmt.Sprintf("[config:struct:11] incorrect type conversion for string field %s", f.field.Name))
			}
			f.parentStruct.FieldByName(f.field.Name).SetString(strval)
			f.fieldType = reflect.String
		}
	}

}

var (
	defaultPath       string = *flag.String("config.path", utils.GetWD(), "path to the configuration file")
	defaultFormat     string = *flag.String("config.format", "json", "format of the configuration file")
	defaultFileName   string = *flag.String("config.filename", "config."+defaultFormat, "name of the configuration file")
	defaultFileSource string = *flag.String("config.source", "", "path to configuration sources")
)

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

func NewConfiguration(options *Options) *Configuration {

	cm := &Configuration{}

	cm.c.init(options)

	return cm
}

func (cm *configuration) init(options *Options) {
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

	cm.properties = make(map[string]*Parameter, initMapLen)
	cm.reader = opt.Reader
	cm.parser = opt.Parser
	cm.source = opt.Source
	cm.envVars = envvar.NewEnvVariables()

}

// Add or modify parameter value
func (cm *configuration) setParameterValue(key string, value any) {
	v := value

	if v == nil {
		s := cm.envVars.GetValueForPropertyKey(key)
		if s == "" {
			return
		}
		v = utils.StringValueToTypedValue(s)
	}

	var (
		cv *Parameter
		ok bool
	)

	cv, ok = cm.properties[key]

	if !ok {
		cv = cm.add(key, utils.TypeSimplified(reflect.TypeOf(v).Kind()))
	}

	cv.SetValue(v)
}

// Add new parameter with nil value.
// Value must be replaced by environment value if it defined.
func (cm *Configuration) Add(key string, valueType ValueDataType) *Parameter {

	return cm.c.add(key, valueType.Kind())
}

func (cm *configuration) add(key string, valueType reflect.Kind) *Parameter {
	if p, ok := cm.properties[key]; ok {
		if p.valueType != valueType {
			panic("[config:defParameter:001] parameter exists with incompatible data types")
		}
		return p
	}

	p := &Parameter{
		name:               key,
		valueType:          valueType,
		value:              nil,
		bindedStructFields: make([]*fieldListItem, 0, initArrayLen),
	}

	cm.replaceByEnvValues(p)

	cm.properties[key] = p

	return p
}

func (cm *Configuration) Set(key string, value any) {
	if key == "" {
		panic("[config:set:01] key is empty")
	}

	if value == nil {
		panic("[config:set:02] value is empty")
	}

	cm.c.setParameterValue(key, value)
}

func (cm *configuration) apply() {

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
	cm.parseEnvPlaceholders(props)

	// Perform correction of numeric values - converting them to int64 or float64 types
	// For the case when all numeric values are converted to one type when reading from the configuration source (e.g., to float64 during JSON deserialization), the correct type is determined by value during reading (reader's Read method).
	correctNumberValues(props)

	// Fill cm.properties with values from received properties
	for k, v := range props {
		cm.setParameterValue(k, v)
	}
}

// Load and apply the configuration from source
func (cm *Configuration) Apply() {
	cm.c.apply()
}

// Replaces value if the configuration parameter value was specified in format ${env_value:default_value}
func (cm *configuration) parseEnvPlaceholders(props map[string]any) {
	for k, v := range props {
		switch str := v.(type) {
		case string:
			{
				pc := strings.Count(str, "${")
				if pc > 0 {
					if pc == 1 && strings.HasPrefix(str, "${") && strings.HasSuffix(str, "}") {
						props[k] = cm.parseValuePlaceholder(str)
						continue
					}

					values := extractPlaceholders(str)
					for _, v := range values {
						value := cm.parseValuePlaceholder(v)
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

func (cm *configuration) parseValuePlaceholder(str string) any {
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

	paramValue := cm.envVars.GetValue(strings.ToUpper(envValue))
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
func (cm *configuration) replaceByEnvValues(value *Parameter) {

	envVar := strings.ReplaceAll(value.name, ".", "_")
	envVal := cm.envVars.GetValue(strings.ToUpper(envVar))
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
	return cm.c.properties[key]
}

// Get all parameters
func (cm *Configuration) GetAll() (parameters []*Parameter) {
	parameters = make([]*Parameter, 0, len(cm.c.properties))
	for _, v := range cm.c.properties {
		parameters = append(parameters, v)
	}
	return
}

// Get parameters by prefix
func (cm *Configuration) Lookup(filter string) (parameters []*Parameter) {
	parameters = make([]*Parameter, 0, len(cm.c.properties))
	for _, v := range cm.c.properties {
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
	cm.c.bind(object)
}

func (cm *configuration) bind(object any) {
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
						fieldType:    utils.TypeSimplified(f.Type.Kind()),
					},
				)
			}
		}
	}
}

// Getting fields of nested structure
func (cm *configuration) collectStructFields(structValue reflect.Value, structField reflect.StructField) {
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
					fieldType:    utils.TypeSimplified(f.Type.Kind()),
				})
			}
		}
	}
}

// Check struct field for config tag and add field to list of bound fields for subsequent filling after configuration reading
func (cm *configuration) bindStructField(f *fieldListItem) {
	tagValue := f.field.Tag.Get("config")
	if tagValue != "" {
		// Check that existing property has same data type as struct attribute
		if p, ok := cm.properties[tagValue]; ok {
			p.bindStructField(f)
			return
		}

		// Add new property
		p := cm.add(tagValue, f.fieldType)
		p.bindStructField(f)
	}
}

func init() {
	parsers = make(map[string]Parser, 3)
	RegisterParser("json", &jsonparser.JsonConfigurationParser{})
	RegisterParser("yaml", &yamlparser.YamlConfigurationParser{})
	RegisterParser("toml", &tomlparser.TomlConfigurationParser{})
}
