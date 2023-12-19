package internal

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"

	log "github.com/sirupsen/logrus"
)

func openBrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Error("Error : ", err)
	}

}

func GetBinaryDir() string {
	if runtime.GOOS == "windows" {
		return "C:\\Cognite\\EdgeExtractor"
	}
	currentDir, _ := os.Getwd()
	return currentDir
}

func EncryptString(key, text string) (string, error) {
	// Convert key to 32 bytes
	keyBytes := []byte(key)
	if len(keyBytes) != 32 {
		return "", errors.New("key must be 32 bytes")
	}

	// Convert text to bytes
	textBytes := []byte(text)

	// Generate a new AES cipher using the key
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	//Create a new GCM - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	//https://golang.org/pkg/crypto/cipher/#NewGCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	//Create a nonce. Nonce should be from GCM
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}

	//Encrypt the data using aesGCM.Seal
	ciphertext := aesGCM.Seal(nonce, nonce, textBytes, nil)
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func DecryptString(key, text string) (string, error) {
	// Convert key to 32 bytes
	keyBytes := []byte(key)
	if len(keyBytes) != 32 {
		return "", errors.New("key must be 32 bytes")
	}

	// Convert text to bytes
	textBytes, err := base64.URLEncoding.DecodeString(text)
	if err != nil {
		return "", err
	}

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	//Create a new GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	//Get the nonce size
	nonceSize := aesGCM.NonceSize()

	//Extract the nonce from the encrypted data
	nonce, ciphertext := textBytes[:nonceSize], textBytes[nonceSize:]

	//Decrypt the data
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
