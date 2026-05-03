package types

type Parser interface {
	Parse(data []byte) (map[string]any, error)
}

var parsers map[string]Parser

func init() {
	parsers = make(map[string]Parser, 3)
}

func RegisterParser(format string, parser Parser) {
	if format == "" {
		panic("[config:reg:01] format must be defined")
	}

	if parser == nil {
		panic("[config:reg:02] parser must be defined")
	}

	parsers[format] = parser
}

func GetParserByFormat(format string) (Parser, bool) {
	parser, ok := parsers[format]
	return parser, ok
}
