package obfuscator

import (
	"bytes"
	"testing"
)

// TestXorEncrypt проверяет, что двойное XOR-шифрование с одним и тем же ключом
// возвращает исходные данные. Это основное свойство алгоритма.
func TestXorEncrypt(t *testing.T) {
	// Определяем тестовые данные и ключ
	originalData := []byte("hello, world!")
	key := byte(0xAB)

	// 1. Шифруем данные
	encryptedData := xorEncrypt(originalData, key)

	// Проверяем, что зашифрованные данные не совпадают с исходными
	if bytes.Equal(originalData, encryptedData) {
		t.Errorf("Encrypted data should not be the same as original data")
	}

	// 2. Расшифровываем данные (применяем XOR еще раз)
	decryptedData := xorEncrypt(encryptedData, key)

	// 3. Сравниваем результат с исходными данными
	if !bytes.Equal(originalData, decryptedData) {
		t.Errorf("Decrypted data does not match original data. Got %q, want %q", decryptedData, originalData)
	}
}

// TestXorEncryptWithZeroKey проверяет поведение с нулевым ключом.
// В этом случае данные не должны изменяться.
func TestXorEncryptWithZeroKey(t *testing.T) {
	originalData := []byte("another test case")
	key := byte(0x00)

	encryptedData := xorEncrypt(originalData, key)

	if !bytes.Equal(originalData, encryptedData) {
		t.Errorf("XOR with zero key should not change the data. Got %q, want %q", encryptedData, originalData)
	}
}
