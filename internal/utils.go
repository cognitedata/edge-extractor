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
	"path/filepath"
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
	filename := os.Args[1] // get command line first parameter
	return filepath.Dir(filename)
}

func EncryptString(key, text string) (string, error) {
	// Convert key to 16 bytes
	keyBytes := []byte(key)
	if len(keyBytes) != 16 {
		return "", errors.New("key must be 16 bytes")
	}

	// Convert text to bytes
	textBytes := []byte(text)

	// Generate a new AES cipher using the key
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	// Generate a new random IV
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	// Encrypt the text using the AES cipher and IV
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(textBytes, textBytes)

	// Combine the IV and encrypted text into a single string
	encrypted := append(iv, textBytes...)
	return base64.URLEncoding.EncodeToString(encrypted), nil
}

func DecryptString(key, text string) (string, error) {
	// Convert key to 16 bytes
	keyBytes := []byte(key)
	if len(keyBytes) != 16 {
		return "", errors.New("key must be 16 bytes")
	}

	// Convert text to bytes
	textBytes, err := base64.URLEncoding.DecodeString(text)
	if err != nil {
		return "", err
	}

	// Generate a new AES cipher using the key
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	// Separate the IV and encrypted text
	iv := textBytes[:aes.BlockSize]
	textBytes = textBytes[aes.BlockSize:]

	// Decrypt the text using the AES cipher and IV
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(textBytes, textBytes)

	return string(textBytes), nil
}
