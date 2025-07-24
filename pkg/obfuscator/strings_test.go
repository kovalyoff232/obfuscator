package obfuscator

import (
	"bytes"
	"testing"
)

func TestXorEncrypt(t *testing.T) {
	originalData := []byte("hello, world!")
	key := byte(0xAB)

	encryptedData := xorEncrypt(originalData, key)

	if bytes.Equal(originalData, encryptedData) {
		t.Errorf("Encrypted data should not be the same as original data")
	}

	decryptedData := xorEncrypt(encryptedData, key)

	if !bytes.Equal(originalData, decryptedData) {
		t.Errorf("Decrypted data does not match original data. Got %q, want %q", decryptedData, originalData)
	}
}

func TestXorEncryptWithZeroKey(t *testing.T) {
	originalData := []byte("another test case")
	key := byte(0x00)

	encryptedData := xorEncrypt(originalData, key)

	if !bytes.Equal(originalData, encryptedData) {
		t.Errorf("XOR with zero key should not change the data. Got %q, want %q", encryptedData, originalData)
	}
}
