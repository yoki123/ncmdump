package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/yoki123/ncmdump"
)

func processFile(name string) {
	defer func() {
		err := recover()
		if err != nil {
			log.Printf("Error processing file:\t%s\n", name)
			log.Printf("Error information:\t\t%v\n", err)
		}
	}()

	fp, err := os.Open(name)
	if err != nil {
		panic(err)
		return
	}
	defer fp.Close()

	if meta, err := ncmdump.DumpMeta(fp); err != nil {
		panic(err)
	} else {
		if data, err := ncmdump.Dump(fp); err != nil {
			panic(err)
		} else {
			log.Printf("Successfully processed:\t%s\n", name)
			output := strings.Replace(name, ".ncm", "."+meta.Format, -1)
			if err = ioutil.WriteFile(output, data, 0644); err != nil {
				panic(err)
			} else {
				log.Printf("Successfully saved file:\t%s\n", output)
				if cover, err := ncmdump.DumpCover(fp); err != nil {
					panic(err)
				} else {
					var tagErr error
					switch meta.Format {
					case "mp3":
						tagErr = addMP3Tag(output, cover, &meta)
					case "flac":
						tagErr = addFLACTag(output, cover, &meta)
					}
					if tagErr != nil {
						log.Printf("Error tagging file:\t%s\n", output)
					} else {
						log.Printf("Successfully tagged file:\t%s\n", output)
					}
				}
			}
		}
	}
}

func main() {
	argc := len(os.Args)
	if argc <= 1 {
		fmt.Println("Please input file path !")
		return
	}
	files := make([]string, 0)

	for i := 0; i < argc-1; i++ {
		path := os.Args[i+1]
		if info, err := os.Stat(path); err != nil {
			log.Printf("Path %s does not exist.", path)
		} else if info.IsDir() {
			filelist, err := ioutil.ReadDir(path)
			if err != nil {
				log.Printf("Error while reading %s: %s", path, err.Error())
			}
			for _, f := range filelist {
				files = append(files, filepath.Join(path, "./", f.Name()))
			}
		} else {
			files = append(files, path)
		}
	}

	for _, filename := range files {
		if filepath.Ext(filename) == ".ncm" {
			processFile(filename)
		}
	}
}
