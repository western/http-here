package crypt

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
)

// ---------------------------------------------------------------------------------------------------------------------------

// PKCS7Padding adds PKCS#7 padding to data
func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// PKCS7Unpadding removes PKCS#7 padding from data
func PKCS7Unpadding(plantext []byte) []byte {
	length := len(plantext)
	unpadding := int(plantext[length-1])
	return plantext[:(length - unpadding)]
}

func CBCEncryptFile(key []byte, inputFile, outputFile string) error {
	plaintext, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("reading input file: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("creating AES cipher: %w", err)
	}

	// Pad the plaintext
	paddedPlaintext := PKCS7Padding(plaintext, aes.BlockSize)

	// Generate a random IV (16 bytes for AES block size)
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return err
	}

	// Create CBC encrypter
	mode := cipher.NewCBCEncrypter(block, iv)

	// Encrypt the data
	ciphertext := make([]byte, len(paddedPlaintext))
	mode.CryptBlocks(ciphertext, paddedPlaintext)

	// Write IV and ciphertext to output file
	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	// Prepend the IV to the ciphertext
	_, err = f.Write(append(iv, ciphertext...))
	return err
}

func CBCDecryptFile(key []byte, inputFile, outputFile string) error {
	ciphertextWithIV, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("reading input file: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("creating AES cipher: %w", err)
	}

	// The IV is the first block size bytes of the ciphertext
	if len(ciphertextWithIV) < aes.BlockSize {
		return fmt.Errorf("ciphertext too short")
	}
	iv := ciphertextWithIV[:aes.BlockSize]
	ciphertext := ciphertextWithIV[aes.BlockSize:]

	// Check if ciphertext is a multiple of the block size (CBC requirement)
	if len(ciphertext)%aes.BlockSize != 0 {
		return fmt.Errorf("ciphertext is not a multiple of the block size")
	}

	// Create CBC decrypter
	mode := cipher.NewCBCDecrypter(block, iv)

	// Decrypt the data
	decrypted := make([]byte, len(ciphertext))
	mode.CryptBlocks(decrypted, ciphertext)

	// Remove padding
	unpaddedPlaintext := PKCS7Unpadding(decrypted)

	// Write the unpadded plaintext to the output file
	return os.WriteFile(outputFile, unpaddedPlaintext, 0644)
}

// ---------------------------------------------------------------------------------------------------------------------------

// (See implementation details for PKCS7Padding and PKCS7UnPadding in referenced docs)

func CBCEncryptHexFile(inputFile, outputFile string, key []byte) error {
	// 1. Read input file
	plaintext, _ := os.ReadFile(inputFile)

	// 2. Pad data, create AES block, generate random IV
	block, _ := aes.NewCipher(key)
	padded := PKCS7Padding(plaintext, aes.BlockSize)
	ciphertext := make([]byte, aes.BlockSize+len(padded))
	iv := ciphertext[:aes.BlockSize]
	io.ReadFull(rand.Reader, iv)

	// 3. Encrypt
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], padded)

	// 4. Encode to hex and write
	return os.WriteFile(outputFile, []byte(hex.EncodeToString(ciphertext)), 0644)
}

func CBCDecryptHexFile(inputFile, outputFile string, key []byte) error {
	// 1. Read and decode hex
	encoded, _ := os.ReadFile(inputFile)
	ciphertext, _ := hex.DecodeString(string(encoded))

	// 2. Extract IV and decrypt
	block, _ := aes.NewCipher(key)
	iv := ciphertext[:aes.BlockSize]
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext[aes.BlockSize:], ciphertext[aes.BlockSize:])

	// 3. Remove padding and write
	return os.WriteFile(outputFile, PKCS7Unpadding(ciphertext[aes.BlockSize:]), 0644)
}

// ---------------------------------------------------------------------------------------------------------------------------
