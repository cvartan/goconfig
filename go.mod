module github.com/cvartan/goconfig

go 1.23.2

require (
	github.com/pelletier/go-toml/v2 v2.2.4
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/testify v1.11.1
)

retract (
	v1.1.1
	v1.1.0
	v1.0.0
	v0.0.2
	v0.0.1
)
