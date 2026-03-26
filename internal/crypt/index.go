package crypt

import (
	"os"

	"crypto/sha256"
	"golang.org/x/crypto/pbkdf2"
)

const (
	defaultIterationCount = 100
	keyLength             = 32
)

func deriveKey(password, salt []byte) []byte {
	// http://www.ietf.org/rfc/rfc2898.txt
	if salt == nil {
		salt = make([]byte, 8)
		// rand.Read(salt)
	}
	return pbkdf2.Key(password, salt, defaultIterationCount, keyLength, sha256.New)
}

func SlurpFile(fileName string) []byte {

	dat, err := os.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	return dat
}

func WriteToFile(fileName string, writeBytes []byte) {

	f, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	_, err = f.Write(writeBytes)
	if err != nil {
		panic(err)
	}
	f.Close()

}
