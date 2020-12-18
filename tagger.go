package ncmdump

import (
	"github.com/bogem/id3v2"
	"github.com/go-flac/flacpicture"
	"github.com/go-flac/flacvorbis"
	"github.com/go-flac/go-flac"
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

type Mp3Tagger struct {
	tag *id3v2.Tag
}

func NewMp3Tagger(path string) (*Mp3Tagger, error) {
	//
	tag, err := id3v2.Open(path, id3v2.Options{Parse: true})
	if err != nil {
		return nil, err
	}
	tagger := new(Mp3Tagger)
	tagger.tag = tag

	return tagger, nil
}

func (m *Mp3Tagger) SetCover(buf []byte, mime string) error {

	m.tag.AddAttachedPicture(id3v2.PictureFrame{
		Encoding:    id3v2.EncodingISO,
		MimeType:    mime,
		PictureType: id3v2.PTFrontCover,
		Description: "Front cover",
		Picture:     buf,
	})
	return nil
}

func (m *Mp3Tagger) SetCoverUrl(coverUrl string) error {

	m.tag.AddAttachedPicture(id3v2.PictureFrame{
		Encoding:    id3v2.EncodingISO,
		MimeType:    "-->",
		PictureType: id3v2.PTFrontCover,
		Description: "Front cover",
		Picture:     []byte(coverUrl),
	})
	return nil
}

func (m *Mp3Tagger) SetTitle(title string) error {
	if name := m.tag.Title(); name == "" {
		m.tag.SetTitle(title)
	}

	return nil
}

//	m.tag.SetDefaultEncoding(id3v2.EncodingUTF8)
func (m *Mp3Tagger) SetAlbum(album string) error {

	if name := m.tag.Album(); name == "" {
		m.tag.SetAlbum(album)
	}
	return nil
}

func (m *Mp3Tagger) SetArtist(artists []string) error {
	// multiple artist support
	if frames := m.tag.GetFrames(m.tag.CommonID("Artist")); len(frames) == 0 {
		for _, artist := range artists {
			m.tag.SetArtist(artist)
		}
	}
	return nil
}

func (m *Mp3Tagger) SetComment(comment string) error {
	if frames := m.tag.GetFrames(m.tag.CommonID("Comments")); len(frames) == 0 {
		m.tag.AddCommentFrame(id3v2.CommentFrame{
			Encoding:    id3v2.EncodingISO,
			Language:    "XXX",
			Description: "",
			Text:        comment,
		})
	}
	return nil
}

func (m *Mp3Tagger) Save() error {
	err := m.tag.Save()
	err = m.tag.Close()
	return err
}

type FlacTagger struct {
	path string
	file *flac.File
	cmts *flacvorbis.MetaDataBlockVorbisComment
}

func NewFlacTagger(path string) (*FlacTagger, error) {
	// already read and closed
	f, err := flac.ParseFile(path)
	if err != nil {
		return nil, err
	}

	// find the vorbisComment
	var cmtmeta *flac.MetaDataBlock
	for _, m := range f.Meta {
		if m.Type == flac.VorbisComment {
			cmtmeta = m
			break
		}
	}
	var cmts *flacvorbis.MetaDataBlockVorbisComment
	if cmtmeta != nil {
		cmts, err = flacvorbis.ParseFromMetaDataBlock(*cmtmeta)
		if err != nil {
			return nil, err
		}
	} else {
		cmts = flacvorbis.New()
	}

	tagger := new(FlacTagger)
	tagger.file = f
	tagger.cmts = cmts
	tagger.path = path
	return tagger, nil
}

func (f *FlacTagger) SetCover(buf []byte, mime string) error {
	picture, err := flacpicture.NewFromImageData(flacpicture.PictureTypeFrontCover, "Front cover", buf, mime)
	if err == nil {
		picturemeta := picture.Marshal()
		f.file.Meta = append(f.file.Meta, &picturemeta)
	}
	return err

}

func (f *FlacTagger) SetCoverUrl(coverUrl string) error {
	picture := &flacpicture.MetadataBlockPicture{
		PictureType: flacpicture.PictureTypeFrontCover,
		MIME:        "-->",
		Description: "Front cover",
		ImageData:   []byte(coverUrl),
	}
	picturemeta := picture.Marshal()
	f.file.Meta = append(f.file.Meta, &picturemeta)
	return nil
}

func (f *FlacTagger) addTag(key string, values ...string) error {
	if old, err := f.cmts.Get(key); err != nil {
		return err
	} else if len(old) == 0 {
		for _, val := range values {
			if err = f.cmts.Add(key, val); err != nil {
				return err
			}
		}
	}
	return nil
}

func (f *FlacTagger) SetTitle(title string) error {
	return f.addTag(flacvorbis.FIELD_TITLE, title)
}

func (f *FlacTagger) SetAlbum(album string) error {

	return f.addTag(flacvorbis.FIELD_ALBUM, album)
}

func (f *FlacTagger) SetArtist(artists []string) error {
	return f.addTag(flacvorbis.FIELD_ARTIST, artists...)
}

// Comment
func (f *FlacTagger) SetComment(comment string) error {
	return f.addTag(flacvorbis.FIELD_DESCRIPTION, comment)
}

func (f *FlacTagger) setVorbisCommentMeta(block *flac.MetaDataBlock) {
	var idx = -1
	for i, m := range f.file.Meta {
		if m.Type == flac.VorbisComment {
			idx = i
			break
		}
	}
	if idx == -1 {
		f.file.Meta = append(f.file.Meta, block)
	} else {

		f.file.Meta[idx] = block
	}
}

func (f *FlacTagger) Save() error {
	block := f.cmts.Marshal()
	f.setVorbisCommentMeta(&block)
	return f.file.Save(f.path)
}
