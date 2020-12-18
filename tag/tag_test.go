package tag

import (
	"testing"
)

func TestNewFlacTagger(t *testing.T) {
	tagger, err := NewFlacTagger("D:\\a.flac")
	if err != nil {
		t.Fatal(err)
	}
	err = tagger.SetComment("11111")
	err = tagger.SetTitle("66666")
	err = tagger.Save()
	t.Log(err)
}
