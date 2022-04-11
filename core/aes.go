package core

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
)

func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

func pkcs7UnPadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("empty data!")
	}
	unPadding := int(data[length-1])
	if unPadding > length {
		return data, nil
	}
	return data[:(length - unPadding)], nil
}

func AESEncrypt(data []byte, tmpkey []byte) (res []byte, err error) {
	key := make([]byte, 32)
	copy(key, tmpkey)
	for len(key) < 32 {
		key = append(key, byte('0'))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	encryptBytes := pkcs7Padding(data, blockSize)
	crypted := make([]byte, len(encryptBytes))
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	blockMode.CryptBlocks(crypted, encryptBytes)

	return []byte(base64.StdEncoding.EncodeToString(crypted)), nil
}

func AESDecrypt(data []byte, tmpkey []byte) (res []byte, err error) {
	key := make([]byte, 32)
	copy(key, tmpkey)
	for len(key) < 32 {
		key = append(key, byte('0'))
	}
	data, err = base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	crypted := make([]byte, len(data))
	blockMode.CryptBlocks(crypted, data)
	crypted, err = pkcs7UnPadding(crypted)
	if err != nil {
		return nil, err
	}
	return crypted, nil
}
