package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"github.com/Reapor-Yurnero/godudect"
	"io"
)

const (
	numberMeasurements = 100000
)

func initState() func([]byte) {
	key := []byte("passphrasewhichneedstobe32bytes!")
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err)
	}
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	return func(plaintext []byte) {
		if len(plaintext)%aes.BlockSize != 0 {
			panic("plaintext is not a multiple of the block size")
		}
		ciphertext := make([]byte, len(plaintext))
		mode.CryptBlocks(ciphertext, plaintext)
	}
}

func prepareInputs() []dudect.Input {
	var inputs = make([]dudect.Input, numberMeasurements)
	for i := 0; i < numberMeasurements; i++ {
		var randByte = make([]byte, 1)
		if n, err := io.ReadFull(rand.Reader, randByte); err != nil || n != 1 {
			panic(fmt.Sprintf("Randbit failed with Err: %v, n: %v", err, n))
		}
		if int(randByte[0])%2 == 0 {
			inputs[i] = dudect.Input{Data: []byte("0000000000000000"), Class: 0}
			continue
		}
		var data = make([]byte, aes.BlockSize)
		if _, err := io.ReadFull(rand.Reader, data); err != nil {
			panic(err)
		}
		inputs[i] = dudect.Input{Data: data, Class: 1}
	}
	return inputs
}

func main() {
	dudect.Dudect(initState, prepareInputs)
}
