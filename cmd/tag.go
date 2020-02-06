package main

import (
	"bytes"
	"github.com/bogem/id3v2"
	"github.com/go-flac/flacpicture"
	"github.com/go-flac/flacvorbis"
	"github.com/go-flac/go-flac"
	"github.com/yoki123/ncmdump"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func containPNGHeader(data []byte) bool {
	if len(data) < 8 {
		return false
	}
	return string(data[:8]) == string([]byte{137, 80, 78, 71, 13, 10, 26, 10})
}

func fetchUrl(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte{}))
	if err != nil {
		return nil, err
	}
	client := http.Client{
		Timeout: 30 * time.Second,
	}
	res, err := client.Do(req)
	if err != nil {

		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		log.Printf("Failed to download album pic: remote returned %d\n", res.StatusCode)
		return nil, err
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {

		return nil, err
	}
	return data, nil
}

func addFLACTag(fileName string, imgData []byte, meta *ncmdump.Meta) {
	f, err := flac.ParseFile(fileName)
	if err != nil {
		log.Println(err)
		return
	}

	if imgData == nil && meta.Album.CoverUrl != "" {
		if coverData, err := fetchUrl(meta.Album.CoverUrl); err != nil {
			log.Println(err)
		} else {
			imgData = coverData
		}
	}

	if imgData != nil {
		picMIME := "image/jpeg"
		if containPNGHeader(imgData) {
			picMIME = "image/png"
		}
		picture, err := flacpicture.NewFromImageData(flacpicture.PictureTypeFrontCover, "Front cover", imgData, picMIME)
		if err == nil {
			picturemeta := picture.Marshal()
			f.Meta = append(f.Meta, &picturemeta)
		}
	} else if meta.Album.CoverUrl != "" {
		picture := &flacpicture.MetadataBlockPicture{
			PictureType: flacpicture.PictureTypeFrontCover,
			MIME:        "-->",
			Description: "Front cover",
			ImageData:   []byte(meta.Album.CoverUrl),
		}
		picturemeta := picture.Marshal()
		f.Meta = append(f.Meta, &picturemeta)
	}

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
			log.Println(err)
			return
		}
	} else {
		cmts = flacvorbis.New()
	}

	if titles, err := cmts.Get(flacvorbis.FIELD_TITLE); err != nil {
		log.Println(err)
		return
	} else if len(titles) == 0 {
		if meta.Name != "" {
			log.Println("Adding music name")
			cmts.Add(flacvorbis.FIELD_TITLE, meta.Name)
		}
	}
	if albums, err := cmts.Get(flacvorbis.FIELD_ALBUM); err != nil {
		log.Println(err)
		return
	} else if len(albums) == 0 {
		if meta.Album.Name != "" {
			log.Println("Adding album name")
			cmts.Add(flacvorbis.FIELD_ALBUM, meta.Album.Name)
		}
	}

	if artists, err := cmts.Get(flacvorbis.FIELD_ARTIST); err != nil {
		log.Println(err)
		return
	} else if len(artists) == 0 {
		for _, artist := range meta.Artists {
			cmts.Add(flacvorbis.FIELD_ARTIST, artist.Name)
		}
	}
	res := cmts.Marshal()
	if cmtmeta != nil {
		*cmtmeta = res
	} else {
		f.Meta = append(f.Meta, &res)
	}

	f.Save(fileName)
}

func addMP3Tag(fileName string, imgData []byte, meta *ncmdump.Meta) {
	tag, err := id3v2.Open(fileName, id3v2.Options{Parse: true})
	if err != nil {
		log.Println(err)
		return
	}
	defer tag.Close()

	if imgData == nil && meta.Album.CoverUrl != "" {
		if coverData, err := fetchUrl(meta.Album.CoverUrl); err != nil {
			log.Println(err)
		} else {
			imgData = coverData
		}
	}

	if imgData != nil {
		picMIME := "image/jpeg"
		if containPNGHeader(imgData) {
			picMIME = "image/png"
		}
		pic := id3v2.PictureFrame{
			Encoding:    id3v2.EncodingISO,
			MimeType:    picMIME,
			PictureType: id3v2.PTFrontCover,
			Description: "Front cover",
			Picture:     imgData,
		}
		tag.AddAttachedPicture(pic)
	} else if meta.Album.CoverUrl != "" {
		pic := id3v2.PictureFrame{
			Encoding:    id3v2.EncodingISO,
			MimeType:    "-->",
			PictureType: id3v2.PTFrontCover,
			Description: "Front cover",
			Picture:     []byte(meta.Album.CoverUrl),
		}
		tag.AddAttachedPicture(pic)
	}

	if tag.GetTextFrame("TIT2").Text == "" {
		if meta.Name != "" {
			log.Println("Adding music name")
			tag.AddTextFrame("TIT2", id3v2.EncodingUTF8, meta.Name)
		}
	}

	if tag.GetTextFrame("TALB").Text == "" {
		if meta.Album.Name != "" {
			log.Println("Adding album name")
			tag.AddTextFrame("TALB", id3v2.EncodingUTF8, meta.Album.Name)
		}
	}

	if tag.GetTextFrame("TPE1").Text == "" {
		for _, artist := range meta.Artists {
			tag.AddTextFrame("TPE1", id3v2.EncodingUTF8, artist.Name)
		}
	}

	if len(tag.GetFrames("COMM")) == 0 {
		if meta.Comment != "" {
			tag.AddCommentFrame(id3v2.CommentFrame{
				Encoding:    id3v2.EncodingISO,
				Language:    "XXX",
				Description: "",
				Text:        meta.Comment,
			})
		}
	}

	if err = tag.Save(); err != nil {
		log.Println(err)
	}
}
