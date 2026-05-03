package types

import "fmt"

type ReadConfigurationSourceError struct {
	parent error
	source string
}

func (e *ReadConfigurationSourceError) Error() string {
	return fmt.Sprintf("can't read configuration source '%s' with error:\n %v", e.source, e.parent)
}

func (e *ReadConfigurationSourceError) Unwrap() error {
	return e.parent
}

func (e *ReadConfigurationSourceError) Is(err error) bool {
	_, ok := err.(*ReadConfigurationSourceError)
	return ok
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

func (e *ParseConfigurationDataError) Unwrap() error {
	return e.parent
}

func (e *ParseConfigurationDataError) Is(err error) bool {
	_, ok := err.(*ParseConfigurationDataError)
	return ok
}

func NewParseConfigurationDataError(parent error, format string) *ParseConfigurationDataError {
	return &ParseConfigurationDataError{
		parent: parent,
	}
}
