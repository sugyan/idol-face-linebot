package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"io"
)

func (a *app) encrypt(src string) (string, error) {
	plain := []byte(src)
	ciphertext := make([]byte, aes.BlockSize+len(plain))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	stream := cipher.NewCFBEncrypter(a.cipherBlock, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plain)
	return hex.EncodeToString(ciphertext), nil
}

func (a *app) decrypt(src string) (string, error) {
	ciphertext, err := hex.DecodeString(src)
	if err != nil {
		return "", err
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(a.cipherBlock, iv)
	stream.XORKeyStream(ciphertext, ciphertext)
	return string(ciphertext), nil
}
