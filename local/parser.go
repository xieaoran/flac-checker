package local

import "github.com/xieaoran/flac-checker/models"

type MetaParser interface {
	Extension() string
	ParseFile(filePath string) (*models.FileMeta, error)
}
