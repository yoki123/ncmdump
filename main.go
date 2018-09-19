package main

import (
	"crypto/aes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"github.com/bogem/id3v2"
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
	for i := len(src) - 1; i >= 0; i-- {
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
	}
	s := l / bs
	return src[:s*bs]
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

func readUint32(rBuf []byte, fp *os.File) uint32 {
	_, err := fp.Read(rBuf)
	if err != nil {
		log.Panic(err)
		return 0
	}
	return binary.LittleEndian.Uint32(rBuf)
}

func processFile(name string) {
	fp, err := os.Open(name)
	defer fp.Close()
	if err != nil {
		log.Println(err)
		return
	}

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
	if err != nil {
		log.Println(err)
		return
	}
	for i := range keyData {
		keyData[i] ^= 0x64
	}

	deKeyData := decryptAes128Ecb(aesCoreKey, fixBlockSize(keyData))
	deKeyData = unPadding(deKeyData)
	// 17 = len("neteasecloudmusic")
	deKeyData = deKeyData[17:]
	for i := range deKeyData {
		c := deKeyData[i]
		if c < 16 {
			deKeyData = deKeyData[:i]
			break
		}
	}

	uLen = readUint32(rBuf, fp)
	var modifyData = make([]byte, uLen)
	_, err = fp.Read(modifyData)
	if err != nil {
		log.Println(err)
		return
	}
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

	// 6 = len("music:")
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

	// crc32 check
	fp.Seek(4, 1)
	fp.Seek(5, 1)

	imgLen := readUint32(rBuf, fp)

	var imgData = make([]byte, imgLen)
	_, err = fp.Read(imgData)
	if err != nil {
		log.Println(err)
		return
	}

	box := buildKeyBox(deKeyData)
	n := 0x8000

	format := "." + musicInfo["format"].(string)
	outputName := strings.Replace(name, ".ncm", format, -1)

	fpOut, err := os.OpenFile(outputName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Println(err)
		return
	}

	var tb = make([]byte, n)
	for {
		_, err := fp.Read(tb)
		if err != nil {
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
		addCover(outputName, imgData)
	} else if format == ".flac" {
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
