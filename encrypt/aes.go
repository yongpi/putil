package encrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

func AESCFBEncrypt(data []byte, key string) ([]byte, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}

	encrypted := make([]byte, aes.BlockSize+len(data))
	iv := encrypted[:aes.BlockSize]

	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(encrypted[aes.BlockSize:], data)

	return encrypted, nil
}

func AESCFBDecrypt(data []byte, key string) ([]byte, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}

	if len(data) < aes.BlockSize {
		return nil, fmt.Errorf("decrypt data too short")
	}

	iv := data[:aes.BlockSize]
	encrypted := data[aes.BlockSize:]
	decrypted := make([]byte, len(encrypted))

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(decrypted, encrypted)

	return decrypted, nil
}

func pkcs5Padding(data []byte, blockSize int) []byte {
	pl := blockSize - len(data)%blockSize
	text := bytes.Repeat([]byte{byte(pl)}, pl)
	data = append(data, text...)

	return data
}

func pkcs5UnPadding(data []byte) []byte {
	length := len(data)
	if length == 0 {
		return nil
	}

	unPadding := int(data[length-1])
	return data[:(length - unPadding)]
}

func AESCBCEncrypt(data []byte, key string) ([]byte, error) {
	lk := len(key)
	if lk != 16 && lk != 24 && lk != 32 {
		return nil, fmt.Errorf("aes cbc key length must be 16 or 24 or 32")
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}

	data = pkcs5Padding(data, block.BlockSize())
	iv := []byte(key[:block.BlockSize()])
	encrypted := make([]byte, len(data))

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(encrypted, data)

	return encrypted, nil
}

func AESCBCDecrypt(data []byte, key string) ([]byte, error) {
	lk := len(key)
	if lk != 16 && lk != 24 && lk != 32 {
		return nil, fmt.Errorf("aes cbc key length must be 16 or 24 or 32")
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}

	iv := []byte(key[:block.BlockSize()])
	decrypted := make([]byte, len(data))

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(decrypted, data)

	decrypted = pkcs5UnPadding(decrypted)
	return decrypted, nil
}
