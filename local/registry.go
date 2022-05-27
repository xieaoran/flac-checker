package local

import (
	"path/filepath"

	"github.com/xieaoran/flac-checker/models"
)

type MetaRegistry struct {
	parsers map[string]MetaParser
}

func NewMetaRegistry() *MetaRegistry {
	registry := &MetaRegistry{parsers: make(map[string]MetaParser)}
	registry.Register(new(FLACMetaParser))
	return registry
}

func (registry *MetaRegistry) Register(parser MetaParser) {
	registry.parsers[parser.Extension()] = parser
}

func (registry *MetaRegistry) ParseFile(filePath string) (*models.FileMeta, bool, error) {
	extension := filepath.Ext(filePath)
	parser, parserFound := registry.parsers[extension]
	if !parserFound {
		return nil, false, nil
	}
	fileMeta, parseError := parser.ParseFile(filePath)
	return fileMeta, true, parseError
}
