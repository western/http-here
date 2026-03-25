package crypt

import (
	"os"
)

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
