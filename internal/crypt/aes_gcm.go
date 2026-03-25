package crypt

import (
	"fmt"
	"io"
	"os"

	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"golang.org/x/crypto/pbkdf2"
)

const (
	defaultIterationCount = 100
	keyLength             = 32
)

var hashFunc = sha256.New
var salt = []byte("salt")

func GCMEncryptFile(password []byte, inputFile, outputFile string) error {

	key := pbkdf2.Key(password, salt, defaultIterationCount, keyLength, hashFunc)

	plaintext, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("reading input file: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("creating AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("creating GCM: %w", err)
	}

	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	// GCM nonce size is typically 12 bytes.
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("generating nonce: %w", err)
	}

	// Seal appends the ciphertext and the authentication tag.
	// It prepends the nonce to the output for easier storage/retrieval.
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	if err := os.WriteFile(outputFile, ciphertext, 0644); err != nil {
		return fmt.Errorf("writing output file: %w", err)
	}

	return nil
}

func GCMDecryptFile(password []byte, inputFile, outputFile string) error {

	key := pbkdf2.Key(password, salt, defaultIterationCount, keyLength, hashFunc)

	data, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("reading input file: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("creating AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("creating GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return fmt.Errorf("ciphertext too short")
	}

	// Extract the nonce from the beginning of the data
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	// Open authenticates and decrypts the data.
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		// This error will occur if the key is wrong or the data has been tampered with.
		return fmt.Errorf("decrypting data: %w", err)
	}

	if err := os.WriteFile(outputFile, plaintext, 0644); err != nil {
		return fmt.Errorf("writing output file: %w", err)
	}

	return nil
}
