package main

import (
	"crypto/aes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-flac/flacpicture"
	"github.com/go-flac/go-flac"

	"github.com/bogem/id3v2"
)

//
var (
	aesCoreKey   = []byte{0x68, 0x7A, 0x48, 0x52, 0x41, 0x6D, 0x73, 0x6F, 0x35, 0x6B, 0x49, 0x6E, 0x62, 0x61, 0x78, 0x57}
	aesModifyKey = []byte{0x23, 0x31, 0x34, 0x6C, 0x6A, 0x6B, 0x5F, 0x21, 0x5C, 0x5D, 0x26, 0x30, 0x55, 0x3C, 0x27, 0x28}
)

func buildKeyBox(key []byte) []byte {
	box := make([]byte, 256)
	for i := 0; i < 256; i++ {
		box[i] = byte(i)
	}
	keyLen := byte(len(key))
	var c, lastByte, keyOffset byte
	for i := 0; i < 256; i++ {
		c = (box[i] + lastByte + key[keyOffset]) & 0xff
		keyOffset++
		if keyOffset >= keyLen {
			keyOffset = 0
		}
		box[i], box[c] = box[c], box[i]
		lastByte = c
	}
	return box
}

func fixBlockSize(src []byte) []byte {
	return src[:len(src)/aes.BlockSize*aes.BlockSize]
}

func PKCS7UnPadding(src []byte) []byte {
	length := len(src)
	unpadding := int(src[length-1])
	return src[:(length - unpadding)]
}

func decryptAes128Ecb(key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	dataLen := len(data)
	decrypted := make([]byte, dataLen)
	bs := block.BlockSize()
	for i := 0; i <= dataLen-bs; i += bs {
		block.Decrypt(decrypted[i:i+bs], data[i:i+bs])
	}
	return PKCS7UnPadding(decrypted), nil
}

func readUint32(rBuf []byte, fp *os.File) uint32 {
	_, err := fp.Read(rBuf)
	checkError(err)
	return binary.LittleEndian.Uint32(rBuf)
}

func checkError(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func processFile(name string) {
	fp, err := os.Open(name)
	if err != nil {
		log.Println(err)
		return
	}
	defer fp.Close()

	var rBuf = make([]byte, 4)
	uLen := readUint32(rBuf, fp)

	if uLen != 0x4e455443 {
		log.Println("isn't netease cloud music copyright file!")
		return
	}

	uLen = readUint32(rBuf, fp)
	if uLen != 0x4d414446 {
		log.Println("isn't netease cloud music copyright file!")
		return
	}

	fp.Seek(2, 1)
	uLen = readUint32(rBuf, fp)

	var keyData = make([]byte, uLen)
	_, err = fp.Read(keyData)
	checkError(err)

	for i := range keyData {
		keyData[i] ^= 0x64
	}

	deKeyData, err := decryptAes128Ecb(aesCoreKey, fixBlockSize(keyData))
	checkError(err)

	// 17 = len("neteasecloudmusic")
	deKeyData = deKeyData[17:]

	uLen = readUint32(rBuf, fp)
	var modifyData = make([]byte, uLen)
	_, err = fp.Read(modifyData)
	checkError(err)

	for i := range modifyData {
		modifyData[i] ^= 0x63
	}
	deModifyData := make([]byte, base64.StdEncoding.DecodedLen(len(modifyData)-22))
	_, err = base64.StdEncoding.Decode(deModifyData, modifyData[22:])
	checkError(err)

	deData, err := decryptAes128Ecb(aesModifyKey, fixBlockSize(deModifyData))
	checkError(err)

	// 6 = len("music:")
	deData = deData[6:]

	var musicInfo map[string]interface{}
	err = json.Unmarshal(deData, &musicInfo)
	checkError(err)

	// crc32 check
	fp.Seek(4, 1)
	fp.Seek(5, 1)

	imgLen := readUint32(rBuf, fp)

	var imgData = make([]byte, imgLen)
	_, err = fp.Read(imgData)
	checkError(err)

	box := buildKeyBox(deKeyData)
	n := 0x8000

	format := "." + musicInfo["format"].(string)
	outputName := strings.Replace(name, ".ncm", format, -1)

	fpOut, err := os.OpenFile(outputName, os.O_RDWR|os.O_CREATE, 0666)
	checkError(err)

	var tb = make([]byte, n)
	for {
		_, err := fp.Read(tb)
		if err != nil { // read EOF
			break
		}
		for i := 0; i < n; i++ {
			j := byte((i + 1) & 0xff)
			tb[i] ^= box[(box[j]+box[(box[j]+j)&0xff])&0xff]
		}
		fpOut.Write(tb)
	}
	fpOut.Close()

	log.Println(outputName)
	if format == ".mp3" {
		addMP3Cover(outputName, imgData)
	} else if format == ".flac" {
		addFLACCover(outputName, imgData)
	}
}

func addFLACCover(fileName string, imgData []byte) {
	f, err := flac.ParseFile(fileName)
	if err != nil {
		log.Println(err)
		return
	}
	picture, err := flacpicture.NewFromImageData(flacpicture.PictureTypeFrontCover, "Front cover", imgData, "image/jpeg")
	if err != nil {
		log.Println(err)
		return
	}
	picturemeta := picture.Marshal()
	f.Meta = append(f.Meta, &picturemeta)
	f.Save(fileName)
}

func addMP3Cover(fileName string, imgData []byte) {
	tag, err := id3v2.Open(fileName, id3v2.Options{Parse: true})
	if err != nil {
		log.Println(err)
		return
	}
	defer tag.Close()

	pic := id3v2.PictureFrame{
		Encoding:    id3v2.EncodingISO,
		MimeType:    "image/jpeg",
		PictureType: id3v2.PTMedia,
		Description: "Front cover",
		Picture:     imgData,
	}

	tag.AddAttachedPicture(pic)

	if err = tag.Save(); err != nil {
		log.Println(err)
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
		} else {
			log.Printf("Skipping %s: not ncm file\n", filename)
		}
	}

}
