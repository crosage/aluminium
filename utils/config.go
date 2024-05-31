package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
)

var JwtKey = "水口钳@jwt"

var AESkey = []byte("水口钳")

func Encrypt(data string) string {
	dataBytes := []byte(data) // 将字符串转换为字节切片
	block, _ := aes.NewCipher(AESkey)
	ciphertext := make([]byte, aes.BlockSize+len(dataBytes))
	iv := ciphertext[:aes.BlockSize]
	if _, err := rand.Read(iv); err != nil {
		panic(err)
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], dataBytes)
	return base64.StdEncoding.EncodeToString(ciphertext)
}

func Decrypt(ciphertext string) string {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		panic(err)
	}
	block, _ := aes.NewCipher(AESkey)
	if len(data) < aes.BlockSize {
		panic("ciphertext too short")
	}
	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(data, data)
	return string(data)
}
