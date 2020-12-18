package ncmdump

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
)

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

func NCMFile(fp *os.File) (bool, error) {
	// Jump to begin of file
	if _, err := fp.Seek(0, io.SeekStart); err != nil {
		return false, err
	}

	var header = make([]byte, 8)
	if err := binary.Read(fp, binary.LittleEndian, &header); err != nil {
		return false, nil
	}

	if string(header) != "CTENFDAM" {
		return false, fmt.Errorf("%s isn't netease cloud music copyright file", fp.Name())
	}

	return true, nil
}

func Decode(fp *os.File) ([]byte, error) {
	// detect whether ncm file
	if result, err := NCMFile(fp); err != nil || !result {
		return nil, err
	}

	// jump over the magic head(4*2) and the gap(2).
	if _, err := fp.Seek(4*2+2, io.SeekStart); err != nil {
		return nil, err
	}

	keyData, err := readLenAndData(fp)
	xorBytes(keyData, 0x64)

	deKeyData, err := decryptAes128Ecb(aesCoreKey, keyData)
	if err != nil {
		return nil, err
	}

	// 17 = len("neteasecloudmusic")
	return deKeyData[17:], nil
}

func DumpMeta(fp *os.File) (Meta, error) {
	// detect whether ncm file
	if result, err := NCMFile(fp); err != nil || !result {
		return Meta{}, err
	}

	// jump over the magic head(4*2) and the gap(2).
	if _, err := fp.Seek(4*2+2, io.SeekStart); err != nil {
		return Meta{}, err
	}

	// whether decode key is successful
	if _, err := Decode(fp); err != nil {
		return Meta{}, err
	}

	modifyData, err := readLenAndData(fp)

	if len(modifyData) == 0 {
		format := "flac"
		if info, err := fp.Stat(); err != nil && info.Size() < int64(math.Pow(1024, 2)*16) {
			format = "mp3"
		}
		return Meta{
			Format: format,
		}, nil
	}

	xorBytes(modifyData, 0x63)

	// 22 = len(`163 key(Don't modify):`)
	deModifyData := make([]byte, base64.StdEncoding.DecodedLen(len(modifyData)-22))
	if _, err = base64.StdEncoding.Decode(deModifyData, modifyData[22:]); err != nil {
		return Meta{}, err
	}

	deData, err := decryptAes128Ecb(aesModifyKey, deModifyData)
	if err != nil {
		return Meta{}, err
	}

	// 6 = len("music:")
	deData = deData[6:]

	var meta Meta
	if err := json.Unmarshal(deData, &meta); err != nil {
		return Meta{}, err
	}

	meta.Comment = string(modifyData)
	return meta, nil
}

func DumpCover(fp *os.File) ([]byte, error) {
	if result, err := NCMFile(fp); !result || err != nil {
		return nil, err
	}

	if _, err := DumpMeta(fp); err != nil {
		return nil, err
	}

	// jump over crc32 check
	if _, err := fp.Seek(9, io.SeekCurrent); err != nil {
		return nil, err
	}

	return readLenAndData(fp)
}

func Dump(fp *os.File) ([]byte, error) {
	if result, err := NCMFile(fp); !result || err != nil {
		return nil, err
	}

	// whether decode key is successful
	deKeyData, err := Decode(fp)
	if err != nil {
		return nil, err
	}

	if _, err := DumpCover(fp); err != nil {
		return nil, err
	}

	box := buildKeyBox(deKeyData)
	n := 0x8000
	var writer bytes.Buffer

	var tb = make([]byte, n)
	for {
		if _, err := fp.Read(tb); err != nil {
			break // read EOF
		}

		for i := 0; i < n; i++ {
			j := byte((i + 1) & 0xff)
			tb[i] ^= box[(box[j]+box[(box[j]+j)&0xff])&0xff]
		}

		writer.Write(tb) // write to memory
	}

	return writer.Bytes(), nil
}
