package tag

import (
	"errors"
	"fmt"
	"strings"
)

const (
	audioFormatMp3  = "mp3"
	audioFormatFlac = "flac"
)

// tag interface for both mp3 and flac
type Tagger interface {
	SetCover(buf []byte, mime string) error // set image buffer
	SetCoverUrl(coverUrl string) error
	SetTitle(string) error
	SetAlbum(string) error
	SetArtist([]string) error
	SetComment(string) error
	Save() error // must be called
}

func NewTagger(input, format string) (Tagger, error) {
	var tagger Tagger
	var err error
	switch strings.ToLower(format) {
	case audioFormatMp3:
		tagger, err = NewMp3Tagger(input)
	case audioFormatFlac:
		tagger, err = NewFlacTagger(input)
	default:
		err = errors.New(fmt.Sprintf("format: %s is not supportted", format))
	}

	return tagger, err
}
