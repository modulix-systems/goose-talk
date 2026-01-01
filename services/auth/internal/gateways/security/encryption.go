package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

func (s *SecurityProvider) EncryptSymmetric(plaintext string, hexKey string) ([]byte, error) {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("security - EncryptSymmetric - hex.DecodeString: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("security - EncryptSymmetric - aes.NewCipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("security - EncryptSymmetric - cipher.NewGCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("security - EncryptSymmetric - io.ReadFull: %w", err)
	}

	encrypted := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return encrypted, nil
}

func (s *SecurityProvider) DecryptSymmetric(encrypted []byte, hexKey string) (string, error) {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return "", fmt.Errorf("security - DecryptSymmetric - hex.DecodeString: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("security - DecryptSymmetric - aes.NewCipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("security - DecryptSymmetric - cipher.NewGCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	nonce, ciphertext := encrypted[:nonceSize], encrypted[nonceSize:]
	decryptedText, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("security - DecryptSymmetric - gcm.Open: %w", err)
	}

	return string(decryptedText), nil
}
