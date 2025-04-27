package hsp

import (
	"bytes"
	"crypto/sha256"
	"testing"
)

func TestGenerateAndDerive(t *testing.T) {
	clientKeys, err := GenerateKeyPair()
	if err != nil {
		t.Fatal("ERR: Failed to generate client keys:", err)
		return
	}

	serverKeys, err := GenerateKeyPair()
	if err != nil {
		t.Fatal("ERR: Failed to generate server keys:", err)
		return
	}

	clientShared, err := DeriveSharedKey(clientKeys.Private, serverKeys.Public)
	if err != nil {
		t.Fatal("ERR: Failed to generate client shared key:", err)
		return
	}

	serverShared, err := DeriveSharedKey(serverKeys.Private, clientKeys.Public)
	if err != nil {
		t.Fatal("ERR: Failed to generate server shared key:", err)
		return
	}

	clientHash := sha256.Sum256(clientShared[:])
	serverHash := sha256.Sum256(serverShared[:])

	t.Logf("Client shared key: %x\n", clientHash)
	t.Logf("Server shared key: %x\n", serverHash)

	// Check they match
	if clientHash == serverHash {
		t.Log("🎉 Secure shared key established! 🎉")
	} else {
		t.Log("❌ Something went wrong ❌")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	clientKeys, err := GenerateKeyPair()
	if err != nil {
		t.Fatal("ERR: Failed to generate client keys:", err)
		return
	}

	serverKeys, err := GenerateKeyPair()
	if err != nil {
		t.Fatal("ERR: Failed to generate server keys:", err)
		return
	}

	clientShared, err := DeriveSharedKey(clientKeys.Private, serverKeys.Public)
	if err != nil {
		t.Fatal("ERR: Failed to generate client shared key:", err)
		return
	}

	serverShared, err := DeriveSharedKey(serverKeys.Private, clientKeys.Public)
	if err != nil {
		t.Fatal("ERR: Failed to generate server shared key:", err)
		return
	}

	clientHash := sha256.Sum256(clientShared[:])
	serverHash := sha256.Sum256(serverShared[:])

	msg := []byte("Hello, World!")

	data, nonce, err := Encrypt(clientHash[:], msg)
	if err != nil {
		t.Error("ERR: Failed to encrypt data:", err)
		return
	}

	decrypted, err := Decrypt(serverHash[:], nonce, data)
	if err != nil {
		t.Error("ERR: Failed to decrypt data:", err)
		return
	}

	if !bytes.Equal(msg, decrypted) {
		t.Error("Plain data doesn't match decrypted one")
	}
}
