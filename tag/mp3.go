package tag

import "github.com/bogem/id3v2"

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
