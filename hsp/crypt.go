package hsp

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	"golang.org/x/crypto/curve25519"
)

type KeyPair struct {
	Private [32]byte
	Public  [32]byte
}

func NewKeyPair(publicKey, privateKey [32]byte) *KeyPair {
	return &KeyPair{
		Public:  publicKey,
		Private: privateKey,
	}
}

func GenerateKeyPair() (pair *KeyPair, err error) {
	privateKey := make([]byte, 32)
	publicKey := make([]byte, 32)
	_, err = rand.Read(privateKey[:])
	if err != nil {
		return
	}

	privateKey[0] &= 248
	privateKey[31] &= 127
	privateKey[31] |= 64

	curve25519.ScalarBaseMult((*[32]byte)(publicKey), (*[32]byte)(privateKey))
	return NewKeyPair([32]byte(publicKey), [32]byte(privateKey)), nil
}

func DeriveSharedKey(privateKey, peerPublicKey [32]byte) (sharedKey [32]byte, err error) {
	generated, err := curve25519.X25519(privateKey[:], peerPublicKey[:])
	if err != nil {
		return
	}

	sharedKey = [32]byte(generated)

	return
}

func Encrypt(key []byte, data []byte) (encrypted []byte, nonce []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	nonce = make([]byte, aesGCM.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, nil, err
	}

	encrypted = aesGCM.Seal(nil, nonce, data, nil)

	return encrypted, nonce, nil
}

func Decrypt(key []byte, nonce []byte, encrypted []byte) (data []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	data, err = aesGCM.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return nil, err
	}

	return data, nil
}
