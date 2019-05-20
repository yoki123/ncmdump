package ncmdump

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
)

var MusicDirectory = os.Getenv("HOME") + "/Music/网易云音乐"

func readfiles(path string) []string {
	var files []string
	filelist, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatalf("Error while reading %s: %s", path, err.Error())
	}
	for _, f := range filelist {
		files = append(files, filepath.Join(path, "./", f.Name()))
	}

	return files
}

func TestNCMFile(t *testing.T) {
	fp, err := os.Open("/etc/hosts")
	if err != nil {
		log.Println(err)
		return
	}
	defer fp.Close()

	if _, err := NCMFile(fp); err != nil {
		t.Log(err)
	} else {
		t.Fatal("not a ncm file, but detect is ok")
	}

	for _, k := range readfiles(MusicDirectory) {
		if filepath.Ext(k) == ".ncm" {

			fp, err := os.Open(k)
			if err != nil {
				log.Println(err)
				return
			}
			defer fp.Close()

			if _, err := NCMFile(fp); err != nil {
				t.Fatal(err)
			} else {
				t.Log(k + " is a ncm file")
			}
		}
	}
}

func TestDumpMeta(t *testing.T) {
	for _, k := range readfiles(MusicDirectory) {
		if filepath.Ext(k) == ".ncm" {

			fp, err := os.Open(k)
			if err != nil {
				log.Println(err)
				return
			}
			defer fp.Close()

			if meta, err := DumpMeta(fp); err != nil {
				t.Fatal(err)
			} else {
				switch meta.Format {
				case "mp3":
					t.Log("file format is mp3")
					if len(meta.Name) <= 0 {
						t.Fatal("cloud not get song's name")
					}

				case "flac":
					t.Log("file format is flac")

				default:
					t.Fatal("unknown format " + meta.Format)
				}
			}
		}
	}
}

func TestDumpCover(t *testing.T) {
	for _, k := range readfiles(MusicDirectory) {
		if filepath.Ext(k) == ".ncm" {
			fp, err := os.Open(k)
			if err != nil {
				log.Println(err)
				return
			}
			defer fp.Close()

			if cover, err := DumpCover(fp); err != nil {
				t.Fatal(err)
			} else {
				if len(cover) > 0 {
					t.Log("get cover from file is success")
				}
			}
		}
	}
}

func TestDecode(t *testing.T) {
	for _, k := range readfiles(MusicDirectory) {
		if filepath.Ext(k) == ".ncm" {
			fp, err := os.Open(k)
			if err != nil {
				log.Println(err)
				return
			}
			defer fp.Close()

			if code, err := Decode(fp); err != nil {
				t.Fatal(err)
			} else {
				if len(code) > 0 {
					t.Log("decode file is success")
				} else {
					t.Fatal("decode file failed")
				}
			}
		}
	}
}

func TestDump(t *testing.T) {
	for _, k := range readfiles(MusicDirectory) {
		if filepath.Ext(k) == ".ncm" {
			fp, err := os.Open(k)
			if err != nil {
				log.Println(err)
				return
			}
			defer fp.Close()

			if dump, err := Dump(fp); err != nil {
				t.Fatal(err)
			} else {
				if len(dump) > 0 {
					t.Log("dump file is success")
				} else {
					t.Fatal("dump file failed")
				}
			}
		}
	}
}
