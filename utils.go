package ncmdump

import (
	"crypto/aes"
	"encoding/binary"
	"io"
	"os"
)

func _PKCS7UnPadding(src []byte) []byte {
	length := len(src)
	unpadding := int(src[length-1])
	return src[:(length - unpadding)]
}

func decryptAes128Ecb(key, data []byte) ([]byte, error) {

	data = data[:len(data)/aes.BlockSize*aes.BlockSize] // unpadding for block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	dataLen := len(data)
	decrypted := make([]byte, dataLen)
	bs := block.BlockSize()
	for i := 0; i <= dataLen-bs; i += bs {
		block.Decrypt(decrypted[i:i+bs], data[i:i+bs])
	}
	return _PKCS7UnPadding(decrypted), nil
}

func readUint32(r io.Reader) (res uint32, err error) {
	err = binary.Read(r, binary.LittleEndian, &res)

	return
}

func xorBytes(data []byte, val uint8) {
	for i := range data {
		data[i] ^= val
	}
}

// read len+data
func readLenAndData(fp *os.File) ([]byte, error) {

	dataLen, err := readUint32(fp)
	if err != nil {
		return nil, err
	}
	if dataLen <= 0 {
		return []byte{}, nil
	}

	var data = make([]byte, dataLen)
	if _, err := fp.Read(data); err != nil {
		return nil, err
	}
	return data, nil
}
