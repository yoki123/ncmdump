package main

import (
	"crypto/aes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"github.com/bogem/id3v2"
	"io/ioutil"
	"log"
	"os"
	"strings"
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

//
func unPadding(src []byte) []byte {
	for i := len(src) - 1; ; i-- {
		if src[i] > 16 {
			return src[:i+1]
		}
	}
	return nil
}

func fixBlockSize(src []byte) []byte {
	bs := 16 // 128
	l := len(src)
	y := l % bs
	if y == 0 {
		return src
	} else {
		s := l / bs
		return src[:s*bs]
	}
}

func decryptAes128Ecb(key, data []byte) []byte {
	dataLen := len(data)
	block, _ := aes.NewCipher([]byte(key))
	decrypted := make([]byte, dataLen)
	size := 16

	for bs, be := 0, size; bs < dataLen; bs, be = bs+size, be+size {
		block.Decrypt(decrypted[bs:be], data[bs:be])
	}
	return unPadding(decrypted)
}

func processFile(fileName string) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return
	}
	var curSeek = 0

	uLen := binary.LittleEndian.Uint32(data[curSeek : curSeek+4])
	if uLen != 0x4e455443 {
		log.Println("isn't netease cloud music copyright file!")
		return
	}
	curSeek += 4

	uLen = binary.LittleEndian.Uint32(data[curSeek : curSeek+4])
	curSeek += 4
	if uLen != 0x4d414446 {
		log.Println("isn't netease cloud music copyright file!")
	}

	curSeek += 2

	keyLen := binary.LittleEndian.Uint32(data[curSeek : curSeek+4])
	curSeek += 4

	keyData := data[curSeek : curSeek+int(keyLen)]
	curSeek += int(keyLen)

	for i := range keyData {
		keyData[i] ^= 0x64
	}

	deKeyData := decryptAes128Ecb(aesCoreKey, fixBlockSize(keyData))
	deKeyData = unPadding(deKeyData)

	deKeyData = deKeyData[17:]

	for i := range deKeyData {
		c := deKeyData[i]
		if c < 16 {
			deKeyData = deKeyData[:i]
			break
		}
	}

	modifyDataLen := binary.LittleEndian.Uint32(data[curSeek : curSeek+4])
	curSeek += 4

	modifyData := data[curSeek : curSeek+int(modifyDataLen)]
	curSeek += int(modifyDataLen)
	for i := range modifyData {
		modifyData[i] ^= 0x63
	}

	deModifyData := make([]byte, base64.StdEncoding.DecodedLen(len(modifyData)-22))
	_, err = base64.StdEncoding.Decode(deModifyData, modifyData[22:])
	if err != nil {
		log.Println(err)
		return
	}

	deData := decryptAes128Ecb(aesModifyKey, fixBlockSize(deModifyData))

	deData = deData[6:]
	for i := range deData {
		c := deData[i]
		if c < 16 {
			deData = deData[:i]
			break
		}
	}

	var musicInfo map[string]interface{}
	err = json.Unmarshal(deData, &musicInfo)
	if err != nil {
		log.Println(err)
		return
	}

	curSeek += 4 + 5
	imgLen := binary.LittleEndian.Uint32(data[curSeek : curSeek+4])
	curSeek += 4

	imgData := data[curSeek : curSeek+int(imgLen)]
	curSeek += int(imgLen)

	box := buildKeyBox(deKeyData)
	n := 0x8000
	lenLeft := len(data) - curSeek
	buff := make([]byte, 0, lenLeft)

	for curSeek+n <= len(data) {
		tb := data[curSeek : curSeek+n]
		for i := 0; i < n; i++ {
			j := byte((i + 1) & 0xff)
			tb[i] ^= box[(box[j]+box[(box[j]+j)&0xff])&0xff]
		}
		curSeek += n
		buff = append(buff, tb...)
	}

	format := musicInfo["format"].(string)
	musicFileName := strings.Replace(fileName, ".ncm", "."+format, -1)
	ioutil.WriteFile(musicFileName, buff, os.ModePerm)
	log.Println(musicFileName)
	if format == "mp3" {
		addCover(musicFileName, imgData)
	} else if format == "flac" {
		// TODO
	}
}

func addCover(fileName string, imgData []byte) {
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
	for i := 0; i < argc-1; i++ {
		processFile(os.Args[i+1])
	}
}
