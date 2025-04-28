package goconfig

import (
	"errors"
	"fmt"
	"maps"
	"reflect"
	"strings"
)

type ConfigurationReader interface {
	Read(string) (map[string]any, error)
}

type ConfigurationWriter interface {
	Write(string, map[string]any) error
}

type ConfigurationManager struct {
	properties map[string]any
	paths      []string
	usedsource string
	reader     ConfigurationReader
	writer     ConfigurationWriter
}

func New() *ConfigurationManager {

	cm := &ConfigurationManager{
		properties: make(map[string]any, 32),
		paths:      make([]string, 0, 1),
		reader:     nil,
		writer:     nil,
	}

	return cm
}

func (cm *ConfigurationManager) SetReader(reader ConfigurationReader) *ConfigurationManager {
	if reader == nil {
		panic("configuration reader must be defined")
	}

	cm.reader = reader

	return cm
}

func (cm *ConfigurationManager) SetWriter(writer ConfigurationWriter) *ConfigurationManager {
	if writer == nil {
		panic("configuration writer must be defined")
	}

	cm.writer = writer
	return cm
}

// Указание конфигурационного файла (с путем к нему).
// Путь может быть абсолютным или относительным (к рабочей директории приложения).
// При задании путей можно создавать цепочку методов. Например:
//
//	config := New().SetSource("config.json").SetSource("../config/config.json")
func (cm *ConfigurationManager) SetSource(path string) *ConfigurationManager {
	if path == "" {
		panic("name of configuration file must be defined")
	}

	cm.paths = append(cm.paths, path)

	return cm
}

func (cm *ConfigurationManager) SetDest(path string) *ConfigurationManager {
	cm.usedsource = path

	return cm
}

// Чтение конфигурационного файла
// Если задано несколько путей, то вычитывается первый который есть - так что при добавлении файлов-источников в SetSource надо учитывать последовательность чтения
func (cm *ConfigurationManager) Read() error {
	if cm.reader == nil {
		panic("configuration reader is not defined")
	}
	if cm.usedsource != "" {
		props, err := cm.reader.Read(cm.usedsource)
		if err == nil {
			maps.Copy(cm.properties, props)
			return nil
		}

		return err
	}

	for _, filename := range cm.paths {
		props, err := cm.reader.Read(filename)
		if err == nil {
			maps.Copy(cm.properties, props)
			cm.usedsource = filename
			return nil
		}

	}

	return errors.New("configuration has not been read")
}

func (cm *ConfigurationManager) Save() error {
	if cm.writer == nil {
		panic("configuration writer is not defined")
	}
	if cm.usedsource == "" {
		panic("configuration destination is not defined")
	}

	return cm.writer.Write(cm.usedsource, cm.properties)
}

// Заполнение структуры данными свойств из конфигурации
// Для того, чтобы атрибут структуры был заполнен надо:
// 1. атрибут должен быть публичным (то есть с большой буквы)
// 2. атрибут должен иметь тэг config в котором указано имя связанного свойства
//
// Например:
//
//	type Config struct {
//	    Attr1 string `config:"app.attr1"`
//	    ...
//	}
func (cm *ConfigurationManager) FillConfig(config any) (err error) {
	c := reflect.ValueOf(config).Elem()
	if c.Type().Kind() != reflect.Struct {
		return fmt.Errorf("filling config must be struct")
	}

	// Собираем все поля, в том числе и во вложенных структурах
	v := make([]fieldListItem, 0, c.NumField())

	for _, f := range reflect.VisibleFields(c.Type()) {
		switch f.Type.Kind() {
		case reflect.Struct:
			{
				buf := collectStructFields(c.FieldByName(f.Name), f)
				v = append(v, buf...)
			}
		case reflect.Array, reflect.Slice:
			{
				continue
			}
		default:
			{
				v = append(v, fieldListItem{
					value: c,
					field: f,
				})
			}
		}
	}

	// Бежим по полям и у тех у кого есть тэг config устанавливаем значение из массива свойств
	for _, f := range v {
		tagValue := f.field.Tag.Get("config")
		if propValue, ok := cm.properties[tagValue]; ok {
			switch reflect.TypeOf(propValue).Kind() {
			case reflect.Int:
				{
					intval, ok := propValue.(int)
					if !ok {
						return fmt.Errorf("incorrect type conversion for integer field %s", f.field.Name)
					}
					f.value.FieldByName(f.field.Name).SetInt(int64(intval))
				}
			case reflect.Bool:
				{
					boolval, ok := propValue.(bool)
					if !ok {
						return fmt.Errorf("incorrect type conversion for boolean field %s", f.field.Name)
					}
					f.value.FieldByName(f.field.Name).SetBool(boolval)
				}
			case reflect.String:
				{
					strval, ok := propValue.(string)
					if !ok {
						return fmt.Errorf("incorrect type conversion for string field %s", f.field.Name)
					}
					f.value.FieldByName(f.field.Name).SetString(strval)
				}
			}
		}
	}
	return
}

type fieldListItem struct {
	value reflect.Value       // Структура, к которой относится поле
	field reflect.StructField // Поле
}

func collectStructFields(structValue reflect.Value, structField reflect.StructField) []fieldListItem {
	result := make([]fieldListItem, 0, 16)
	fields := reflect.VisibleFields(structField.Type)
	for _, f := range fields {
		switch f.Type.Kind() {
		case reflect.Struct:
			{
				innerFields := collectStructFields(structValue.FieldByName(f.Name), f)
				result = append(result, innerFields...)
			}
		case reflect.Array, reflect.Slice:
			{
				// TODO: видно когда-то надо придумать как работать с массивами
				continue
			}
		default:
			{
				result = append(result, fieldListItem{
					value: structValue,
					field: f,
				})
			}
		}
	}

	return result
}

func (cm *ConfigurationManager) Set(propertyName string, value any) {
	cm.properties[propertyName] = value
}

func (cm *ConfigurationManager) Get(propertyName string) any {
	return cm.properties[propertyName]
}

func (cm *ConfigurationManager) Delete(propertyName string) {
	delete(cm.properties, propertyName)
}

func (cm *ConfigurationManager) GetAll() map[string]any {
	return maps.Clone(cm.properties)
}

func (cm *ConfigurationManager) Lookup(param string) (result map[string]any) {
	result = make(map[string]any, len(cm.properties))

	for k, v := range cm.properties {
		if strings.HasPrefix(k, param) {
			result[k] = v
		}
	}

	return
}
