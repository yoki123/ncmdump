package main

import (
	"fmt"
	"github.com/mingcheng/ncmdump.go"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func processFile(name string) {
	fp, err := os.Open(name)
	if err != nil {
		log.Println(err)
		return
	}
	defer fp.Close()

	if meta, err := ncmdump.DumpMeta(fp); err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("%s: %s\n", name, meta.Format)
	}
}

func main() {
	argc := len(os.Args)
	if argc <= 1 {
		log.Println("please input file path!")
		return
	}
	files := make([]string, 0)

	for i := 0; i < argc-1; i++ {
		path := os.Args[i+1]
		if info, err := os.Stat(path); err != nil {
			log.Fatalf("Path %s does not exist.", info)
		} else if info.IsDir() {
			filelist, err := ioutil.ReadDir(path)
			if err != nil {
				log.Fatalf("Error while reading %s: %s", path, err.Error())
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
