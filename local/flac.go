package local

import (
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/go-flac/flacvorbis"
	"github.com/go-flac/go-flac"
	"github.com/xieaoran/flac-checker/models"
)

type FLACMetaParser struct {
}

func (*FLACMetaParser) Extension() string {
	return ".flac"
}

func (*FLACMetaParser) ParseFile(filePath string) (*models.FileMeta, error) {
	fileMeta := &models.FileMeta{FilePath: filePath}

	file, fileError := flac.ParseFile(filePath)
	if fileError != nil {
		return fileMeta, fileError
	}
	streamInfo, streamError := file.GetStreamInfo()
	if streamError != nil {
		return fileMeta, streamError
	}
	fileMeta.BitDepth = streamInfo.BitDepth
	fileMeta.SampleRate = streamInfo.SampleRate
	fileMeta.AudioMD5 = hex.EncodeToString(streamInfo.AudioMD5)

	metaExists := false
	for _, metaBlock := range file.Meta {
		if metaBlock.Type != flac.VorbisComment {
			continue
		}
		metaExists = true

		comment, commentError := flacvorbis.ParseFromMetaDataBlock(*metaBlock)
		if commentError != nil {
			return fileMeta, commentError
		}
		titles, titleError := comment.Get(flacvorbis.FIELD_TITLE)
		if titleError != nil {
			return fileMeta, titleError
		}
		if len(titles) == 0 {
			return fileMeta, fmt.Errorf("TITLE empty, comment.Comments[%+v]", comment.Comments)
		}
		artists, artistError := comment.Get(flacvorbis.FIELD_ARTIST)
		if artistError != nil {
			return fileMeta, artistError
		}
		if len(artists) == 0 {
			return fileMeta, fmt.Errorf("ARTIST empty, comment.Comments[%+v]", comment.Comments)
		}
		albums, albumError := comment.Get(flacvorbis.FIELD_ALBUM)
		if albumError != nil {
			return fileMeta, albumError
		}
		if len(albums) == 0 {
			return fileMeta, fmt.Errorf("ALBUM empty, comment.Comments[%+v]", comment.Comments)
		}

		fileMeta.Title = titles[0]
		fileMeta.Artist = artists[0]
		fileMeta.Album = albums[0]
	}

	if !metaExists {
		return fileMeta, errors.New("VorbisComment not exists in file.Meta")
	}
	return fileMeta, nil
}
