package parser

import (
	"fmt"
	"io"
	"strings"
	"sync"
)

var (
	globalRegistry = &Registry{
		parsers: make(map[string]Parser),
	}
	registryMutex sync.RWMutex
)

// Registry holds registered parsers for different formats
type Registry struct {
	parsers map[string]Parser
}

// Register adds a parser for a specific format to the global registry
func Register(format string, parser Parser) {
	registryMutex.Lock()
	defer registryMutex.Unlock()
	globalRegistry.parsers[strings.ToLower(format)] = parser
}

// GetParser returns a parser for the specified format from the global registry
func GetParser(format string) (Parser, error) {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	parser, ok := globalRegistry.parsers[strings.ToLower(format)]
	if !ok {
		return nil, fmt.Errorf("no parser registered for format: %s", format)
	}
	return parser, nil
}

// Parse is a convenience function to parse a file using the global registry
func Parse(format, filePath string) (*Book, error) {
	parser, err := GetParser(format)
	if err != nil {
		return nil, err
	}
	return parser.Parse(filePath)
}

// ParseReader is a convenience function to parse from a reader using the global registry
func ParseReader(format string, r io.ReaderAt, size int64) (*Book, error) {
	parser, err := GetParser(format)
	if err != nil {
		return nil, err
	}
	return parser.ParseReader(r, size)
}

// RegisteredFormats returns a list of all registered format identifiers
func RegisteredFormats() []string {
	registryMutex.RLock()
	defer registryMutex.RUnlock()

	formats := make([]string, 0, len(globalRegistry.parsers))
	for format := range globalRegistry.parsers {
		formats = append(formats, format)
	}
	return formats
}
