// Copyright (c) 2022 Institute of Software, Chinese Academy of Sciences (ISCAS)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package storage

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
	"os"
)

func NewCryptReader(r io.Reader, pass string) io.Reader {
	if len(pass) == 0 {
		return r
	}

	key := append([]byte{}, []byte(pass)...)
	key = append(key, make([]byte, 32)...)

	block, _ := aes.NewCipher(key[:32])

	var iv [aes.BlockSize]byte
	stream := cipher.NewOFB(block, iv[:])
	return &cipher.StreamReader{S: stream, R: r}
}

func NewCryptWriter(w io.Writer, pass string) io.Writer {
	if len(pass) == 0 {
		return w
	}

	key := append([]byte{}, []byte(pass)...)
	key = append(key, make([]byte, 32)...)

	block, _ := aes.NewCipher(key[:32])

	// If the key is unique for each ciphertext, then it's ok to use a zero
	// IV.
	var iv [aes.BlockSize]byte
	stream := cipher.NewOFB(block, iv[:])

	return &cipher.StreamWriter{S: stream, W: w}
}

func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func PKCS7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

//AES加密,CBC
func AesEncrypt(origData []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key[:32])
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = PKCS7Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

//AES解密
func AesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key[:32])
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS7UnPadding(origData)
	return origData, nil
}

func Hash(src []byte, rounds int) []byte {
	dst := src
	for i := 0; i < rounds; i++ {
		t := sha256.Sum256(dst)
		dst = t[0:]
	}
	return dst
}
func HashHex(src []byte, rounds int) string {
	h := Hash(src, rounds)
	return hex.EncodeToString(h)
}

func Md5Sum(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return ""
	}
	return hex.EncodeToString(h.Sum(nil))
}

func AesDecryption(str string, pass string) (string, error) {
	decodeBytes, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return "", err
	}
	decryptBytes, err := AesDecrypt(decodeBytes, []byte(pass))
	if err != nil {
		return "", err
	}
	return string(decryptBytes), nil
}

