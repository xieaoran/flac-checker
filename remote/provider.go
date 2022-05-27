package remote

import "github.com/xieaoran/flac-checker/models"

type OnlineProvider interface {
	Name() string
	Search(localMeta *models.FileMeta) ([]*models.FileMeta, error)
}
