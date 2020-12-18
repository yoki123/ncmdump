package tag

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/yoki123/ncmdump"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	mimeJPEG = "image/jpeg"
	mimePNG  = "image/png"
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
		return nil, errors.New(fmt.Sprintf("Failed to download album pic: remote returned %d\n", res.StatusCode))
	}
	defer res.Body.Close()

	return ioutil.ReadAll(res.Body)
}

func TagAudioFileFromMeta(tag Tagger, imgData []byte, meta *ncmdump.Meta) error {
	if imgData == nil && meta.Album.CoverUrl != "" {
		if coverData, err := fetchUrl(meta.Album.CoverUrl); err != nil {
			log.Println(err)
		} else {
			imgData = coverData
		}
	}

	if imgData != nil {
		picMIME := mimeJPEG
		if containPNGHeader(imgData) {
			picMIME = mimePNG
		}
		tag.SetCover(imgData, picMIME)
	} else if meta.Album.CoverUrl != "" {
		tag.SetCoverUrl(meta.Album.CoverUrl)
	}

	if meta.Name == "" {
		tag.SetTitle(meta.Name)
	}

	if meta.Album.Name == "" {
		tag.SetAlbum(meta.Name)
	}

	artists := make([]string, 0)
	for _, artist := range meta.Artists {
		artists = append(artists, artist.Name)
	}
	if len(artists) > 0 {
		tag.SetArtist(artists)
	}

	if meta.Comment != "" {
		tag.SetComment(meta.Comment)
	}

	return tag.Save()
}
