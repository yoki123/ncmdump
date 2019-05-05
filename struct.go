package ncmdump

import (
	"encoding/json"
)

type Album struct {
	Id       float64 `json:"albumId"`
	Name     string  `json:"album"`
	CoverUrl string  `json:"albumPic"`
}

type Artist struct {
	Name string
	Id   float64
}

// @see https://stackoverflow.com/questions/42377989/unmarshal-json-array-of-arrays-in-go
func (a *Artist) UnmarshalJSON(data []byte) error {
	var v []interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	a.Name = v[0].(string)
	a.Id = v[1].(float64)
	return nil
}

// @ref https://music.163.com/#/song?id={id}
type Meta struct {
	Id       float64  `json:"musicId"`
	Name     string   `json:"musicName"`
	Album    *Album   `json:"-"`
	Artists  []Artist `json:"artist"`
	BitRate  float64  `json:"bitrate"`
	Duration float64  `json:"duration"`
	Format   string   `json:"format"`
}
