package types

import "fmt"

type ReadConfigurationSourceError struct {
	parent error
	source string
}

func (e *ReadConfigurationSourceError) Error() string {
	return fmt.Sprintf("can't read configuration source '%s' with error:\n %v", e.source, e.parent)
}

func NewReadConfigurationSourceError(parent error, source string) *ReadConfigurationSourceError {
	return &ReadConfigurationSourceError{
		source: source,
		parent: parent,
	}
}

type ParseConfigurationDataError struct {
	parent error
	format string
}

func (e *ParseConfigurationDataError) Error() string {
	return fmt.Sprintf("can't parse configuration data from format '%s' with error:\n %v", e.format, e.parent)
}

func NewParseConfigurationDataError(parent error, format string) *ParseConfigurationDataError {
	return &ParseConfigurationDataError{
		parent: parent,
	}
}
