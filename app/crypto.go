package app

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
)

func (app *BotApp) encrypt(src string) (string, error) {
	plain := []byte(src)
	ciphertext := make([]byte, aes.BlockSize+len(plain))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	stream := cipher.NewCFBEncrypter(app.cipherBlock, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plain)
	return base64.RawStdEncoding.EncodeToString(ciphertext), nil
}

func (app *BotApp) decrypt(src string) (string, error) {
	ciphertext, err := base64.RawStdEncoding.DecodeString(src)
	if err != nil {
		return "", err
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]
	stream := cipher.NewCFBDecrypter(app.cipherBlock, iv)
	stream.XORKeyStream(ciphertext, ciphertext)
	return string(ciphertext), nil
}
