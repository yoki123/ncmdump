package main

import (
	"fmt"
	cli "github.com/urfave/cli/v2"
	"github.com/yoki123/ncmdump"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func getOutputFullPath(input string, outputDir string, format string) string {
	if outputDir == "" {
		outputDir = filepath.Dir(input)
	}

	outputDir = filepath.Clean(outputDir)

	name := filepath.Base(input)
	newName := strings.Replace(name, ".ncm", "."+format, -1)
	return outputDir + string(filepath.Separator) + newName
}

func processFile(input string, outputDir string, isTag bool) {
	defer func() {
		err := recover()
		if err != nil {
			log.Printf("Error processing file:\t%s\n", input)
			log.Printf("Error information:\t\t%v\n", err)
		}
	}()
	fp, err := os.Open(input)
	if err != nil {
		panic(err)
	}
	defer fp.Close()

	meta, err := ncmdump.DumpMeta(fp)
	if err != nil {
		panic(err)
	}

	data, err := ncmdump.Dump(fp)
	if err != nil {
		panic(err)
	}

	log.Printf("Successfully processed:\t%s\n", input)

	output := getOutputFullPath(input, outputDir, meta.Format) // #filepathstrings.Replace(input, ".ncm", "."+meta.Format, -1)

	err = ioutil.WriteFile(output, data, 0644)
	if err != nil {
		panic(err)
	}

	log.Printf("Successfully saved file:\t%s\n", output)
	cover, err := ncmdump.DumpCover(fp)
	if err != nil {
		fmt.Println(1, err)
		panic(err)
	}
	if !isTag {
		return
	}

	tagger, err := ncmdump.NewTagger(output, meta.Format)
	if err != nil {
		fmt.Println(2, err)
		panic(err)
	}
	err = ncmdump.TagAudioFileFromMeta(tagger, cover, &meta)
	if err != nil {
		log.Printf("Error tagging file:\t%s\n", output)
	} else {
		log.Printf("Successfully tagged file:\t%s\n", output)
	}
}

func main() {

	app := cli.NewApp()
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "output",
			Value: "",
			Usage: "output directory path.",
		},

		&cli.BoolFlag{
			Name:  "tag",
			Value: true,
			Usage: "tag the output file from ncm file metadata.",
		},
	}

	app.Action = func(c *cli.Context) error {

		args := c.Args().Slice()
		outputDir := c.String("output")
		isTag := c.Bool("tag")

		if len(args) < 1 {
			fmt.Println("Please input file path !")
			return nil
		}

		files := make([]string, 0)
		for _, path := range args {

			if info, err := os.Stat(path); err != nil {
				log.Printf("Path %s does not exist.", path)
			} else if info.IsDir() {
				filelist, err := ioutil.ReadDir(path)
				if err != nil {
					log.Printf("Error while reading %s: %s", path, err.Error())
				}
				for _, f := range filelist {
					files = append(files, filepath.Join(path, string(filepath.Separator), f.Name()))
				}
			} else {
				files = append(files, path)
			}
		}

		for _, filename := range files {
			if strings.ToLower(filepath.Ext(filename)) == ".ncm" {
				processFile(filename, outputDir, isTag)
			}
		}
		return nil
	}
	app.Run(os.Args)
	return
}
