package goconfig_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cvartan/goconfig"
	"github.com/stretchr/testify/assert"
)

type TestConfig struct {
	IntAttr       int     `config:"test.test_int"`
	Int32Attr     int32   `config:"test.test_int"`
	UintAttr      uint    `config:"test.test_int"`
	FloatAttr     float64 `config:"test.test_float"`
	BoolAttr      bool    `config:"test.test_bool"`
	StringAttr    string  `config:"test.test_str"`
	SubStructAttr struct {
		Int16Attr   int16  `config:"test.test_int"`
		BoolAttr    bool   `config:"test.test_bool"`
		StrAttr     string `config:"test.test_str"`
		StrEnvAttr  string `config:"goconfig.strattr"`
		IntEnvAttr  int32  `config:"goconfig.intattr"`
		BoolEnvAttr bool   `config:"goconfig.boolattr"`
	}
}

var jsonConfig string = `
{
    "test":{
        "test_int":100,
		"test_float":100.00,
        "test_bool":true,
        "test_str":"test string",
        "test_array_values":[100,200,300],
		"test_float_array_values":[100.00,200.00,300.00],
        "test_array_obj":[
            {
                "name":"value1",
                "value":100
            },
            {
                "name":"value2",
                "value":200
            },
            {
                "name":"value3",
                "value":300
            }          
        ]
    },
    "env":{
        "strenv":"${GOCONFIG_TEST_STRING:alpha value}",
        "boolenv":"${GOCONFIG_TEST_BOOL}",
        "intenv":"${GOCONFIG_TEST_INT:300}",
        "compenv":"test ${GOCONFIG_TEST_BOOL}"
    }
}
`

var yamlConfig string = `
test:
    test_int: 100
    test_float: 100.00
    test_bool: true
    test_str: test string
    test_array_values: [100, 200, 300]
    test_float_array_values: [100.00, 200.00, 300.00]
    test_array_obj:
        - name: value1
          value: 100
        - name: value2
          value: 200
        - name: value3
          value: 300
env:
    strenv: ${GOCONFIG_TEST_STRING:alpha value}
    boolenv: ${GOCONFIG_TEST_BOOL}
    intenv: ${GOCONFIG_TEST_INT:300}
    compenv: test ${GOCONFIG_TEST_BOOL}
`

var tomlConfig string = `
[test]
	test_int = 100
	test_float = 100.00
	test_bool = true
	test_str = 'test string'
	test_array_values = [100, 200, 300]
	test_float_array_values = [100.00, 200.00, 300.00]
	test_array_obj = [{name = "value1", value = 100},{name = "value2", value = 200},{name = "value3", value = 300}]

[env]
	strenv = '${GOCONFIG_TEST_STRING:alpha value}'
	boolenv = '${GOCONFIG_TEST_BOOL}'
	intenv = '${GOCONFIG_TEST_INT:300}'
	compenv = 'test ${GOCONFIG_TEST_BOOL}'
`

func addEnvVariables() {
	os.Setenv("GOCONFIG_TEST_STRING", "goconfig test string")
	os.Setenv("GOCONFIG_TEST_BOOL", "true")
	os.Setenv("GOCONFIG_TEST_INT", "245")

	os.Setenv("GOCONFIG_STRATTR", "test string")
	os.Setenv("GOCONFIG_INTATTR", "1034")
	os.Setenv("GOCONFIG_BOOLATTR", "true")
}

func generateFileName(ext string) string {
	n := time.Now().Unix()
	return fmt.Sprintf("config-%d.%s", n, ext)
}

func createJsonFile() (string, error) {
	filename := generateFileName("json")
	err := os.WriteFile(filename, []byte(jsonConfig), os.FileMode(0644))
	if err != nil {
		return "", err
	}

	return filename, nil
}

func createYamlFile() (string, error) {
	filename := generateFileName("yml")
	err := os.WriteFile(filename, []byte(yamlConfig), os.FileMode(0644))
	if err != nil {
		return "", err
	}

	return filename, nil
}

func createTomlFile() (string, error) {
	filename := generateFileName("toml")
	err := os.WriteFile(filename, []byte(tomlConfig), os.FileMode(0644))
	if err != nil {
		return "", err
	}

	return filename, nil
}

func deleteFile(filePath string) {
	os.Remove(filePath)
}

/* func shim(a ...any) []any {
	return a
} */

func mustParameter(t *testing.T, value *goconfig.Parameter) *goconfig.Parameter {
	if value != nil {
		return value
	}

	t.Fatal()
	return nil
}

func checkConfig(config *goconfig.Configuration, t *testing.T) {
	assert.Equal(t, int64(100), mustParameter(t, config.Get("test.test_int")).Int())
	assert.Equal(t, float64(100.00), mustParameter(t, config.Get("test.test_float")).Float())
	assert.Equal(t, true, mustParameter(t, config.Get("test.test_bool")).Bool())
	assert.Equal(t, "test string", mustParameter(t, config.Get("test.test_str")).String())
	assert.Equal(t, int64(245), mustParameter(t, config.Get("env.intenv")).Int())
	assert.Equal(t, true, mustParameter(t, config.Get("env.boolenv")).Bool())
	assert.Equal(t, "goconfig test string", mustParameter(t, config.Get("env.strenv")).String())

	assert.Equal(t, int64(100), mustParameter(t, config.Get("test.test_array_values.0")).Int())
	assert.Equal(t, int64(200), mustParameter(t, config.Get("test.test_array_values.1")).Int())
	assert.Equal(t, int64(300), mustParameter(t, config.Get("test.test_array_values.2")).Int())

	values := config.Lookup("test.test_array_values.")

	i0 := values[0].Int()
	i1 := values[1].Int()
	i2 := values[2].Int()

	assert.ElementsMatch(t, []int64{100, 200, 300}, []int64{i0, i1, i2})

	values_arr := config.GetIntArray("test.test_array_values")
	assert.ElementsMatch(t, []int64{100, 200, 300}, values_arr)

	float_values_arr := config.GetFloatArray("test.test_float_array_values")
	assert.ElementsMatch(t, []float64{100.00, 200.00, 300.00}, float_values_arr)

	assert.Equal(t, "value1", mustParameter(t, config.Get("test.test_array_obj.0.name")).String())
	assert.Equal(t, int64(100), mustParameter(t, config.Get("test.test_array_obj.0.value")).Int())
	assert.Equal(t, "value2", mustParameter(t, config.Get("test.test_array_obj.1.name")).String())
	assert.Equal(t, int64(200), mustParameter(t, config.Get("test.test_array_obj.1.value")).Int())
	assert.Equal(t, "value3", mustParameter(t, config.Get("test.test_array_obj.2.name")).String())
	assert.Equal(t, int64(300), mustParameter(t, config.Get("test.test_array_obj.2.value")).Int())
}

func checkStruct(obj *TestConfig, t *testing.T) {
	assert.Equal(t, 100, obj.IntAttr)
	assert.Equal(t, int32(100), obj.Int32Attr)
	assert.Equal(t, uint(100), obj.UintAttr)
	assert.Equal(t, true, obj.BoolAttr)
	assert.Equal(t, "test string", obj.StringAttr)

	assert.Equal(t, int16(100), obj.SubStructAttr.Int16Attr)
	assert.Equal(t, "test string", obj.SubStructAttr.StrAttr)
	assert.Equal(t, true, obj.SubStructAttr.BoolAttr)

	assert.Equal(t, int32(1034), obj.SubStructAttr.IntEnvAttr)
	assert.Equal(t, "test string", obj.SubStructAttr.StrEnvAttr)
	assert.Equal(t, true, obj.SubStructAttr.BoolEnvAttr)
}

func TestReadJsonConfig(t *testing.T) {
	cfile, err := createJsonFile()

	if err != nil {
		t.Fatal(err)
	}
	defer deleteFile(cfile)

	addEnvVariables()

	options := &goconfig.Options{
		Filename: cfile,
		Format:   "json",
	}

	config := goconfig.NewConfiguration(options)

	obj := &TestConfig{}
	config.Bind(obj)
	config.Apply()

	checkConfig(config, t)
	checkStruct(obj, t)

}

func TestReadYamlConfig(t *testing.T) {
	cfile, err := createYamlFile()

	if err != nil {
		t.Fatal(err)
	}
	defer deleteFile(cfile)

	addEnvVariables()

	options := &goconfig.Options{
		Filename: cfile,
		Format:   "yaml",
	}

	config := goconfig.NewConfiguration(options)

	obj := &TestConfig{}
	config.Bind(obj)
	config.Apply()

	checkConfig(config, t)
	checkStruct(obj, t)

}

func TestReadTomlConfig(t *testing.T) {
	cfile, err := createTomlFile()

	if err != nil {
		t.Fatal(err)
	}
	defer deleteFile(cfile)

	addEnvVariables()

	options := &goconfig.Options{
		Filename: cfile,
		Format:   "toml",
	}

	config := goconfig.NewConfiguration(options)

	obj := &TestConfig{}
	config.Bind(obj)
	config.Apply()

	checkConfig(config, t)
	checkStruct(obj, t)
}
