package main

import (
	"errors"
	"fmt"
	cli "github.com/urfave/cli/v2"
	"github.com/yoki123/ncmdump"
	"github.com/yoki123/ncmdump/tag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var VERSION = "VERSION"

// expand tilde `~` to the user's home directory
func expandTilde(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return home + path[1:], nil
}

func mkdirIfNotExist(path string) error {
	info, err := os.Stat(path)

	if os.IsNotExist(err) {
		err = os.MkdirAll(path, 0755)
	}

	if err != nil {
		return err
	}

	if !info.IsDir() {
		return errors.New(fmt.Sprintf("output path is not a directory"))
	}
	return nil
}

func getOutputFullPath(input string, outputDir string, format string) string {
	if outputDir == "" {
		outputDir = filepath.Dir(input)
	} else {
		outputDir = filepath.Clean(outputDir)

		var err error
		if outputDir, err = expandTilde(outputDir); err != nil {
			outputDir = filepath.Dir(input)
			log.Printf("get user's home directory error: %s, write to input path instead\n", err)
		}

		// auto mkdir if not exist
		if err := mkdirIfNotExist(outputDir); err != nil {
			outputDir = filepath.Dir(input)
			log.Printf("stat output path error: %s, write to input path instead\n", err)
		}
	}

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
		panic(err)
	}
	if !isTag {
		return
	}

	tagger, err := tag.NewTagger(output, meta.Format)
	if err != nil {
		panic(err)
	}
	err = tag.TagAudioFileFromMeta(tagger, cover, &meta)
	if err != nil {
		log.Printf("Error tagging file:\t%s\n", output)
	} else {
		log.Printf("Successfully tagged file:\t%s\n", output)
	}
}

func main() {

	app := cli.NewApp()
	app.Version = VERSION
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
