package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

var (
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
	ErrInvalidKey        = errors.New("invalid key length")
)

// GenerateKey generates a random 256-bit (32 byte) key for AES-256-GCM
func GenerateKey() ([]byte, error) {
	key := make([]byte, 32) // AES-256 requires 32 bytes
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}
	return key, nil
}

// GenerateKeyFromPassphrase derives a 256-bit key from a passphrase using SHA-256
func GenerateKeyFromPassphrase(passphrase string) []byte {
	hash := sha256.Sum256([]byte(passphrase))
	return hash[:]
}

// Encrypt encrypts plaintext using AES-256-GCM
// The key must be 32 bytes for AES-256
// Returns base64-encoded ciphertext (nonce + ciphertext)
func Encrypt(plaintext []byte, key []byte) (string, error) {
	if len(key) != 32 {
		return "", ErrInvalidKey
	}

	// Create cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher block: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and authenticate
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Encode to base64
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts ciphertext using AES-256-GCM
// The key must be 32 bytes for AES-256
// Expects base64-encoded ciphertext (nonce + ciphertext)
func Decrypt(ciphertext string, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, ErrInvalidKey
	}

	// Decode from base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid base64", ErrInvalidCiphertext)
	}

	// Create cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher block: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Check minimum length (nonce + ciphertext)
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, ErrInvalidCiphertext
	}

	// Extract nonce and ciphertext
	nonce, ciphertextData := data[:nonceSize], data[nonceSize:]

	// Decrypt and verify
	plaintext, err := gcm.Open(nil, nonce, ciphertextData, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: decryption failed", ErrInvalidCiphertext)
	}

	return plaintext, nil
}

// EncryptString encrypts a string using AES-256-GCM
func EncryptString(plaintext string, key []byte) (string, error) {
	return Encrypt([]byte(plaintext), key)
}

// DecryptString decrypts a string using AES-256-GCM
func DecryptString(ciphertext string, key []byte) (string, error) {
	data, err := Decrypt(ciphertext, key)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Hash generates a SHA-256 hash of the input data
func Hash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// HashString generates a SHA-256 hash of the input string and returns it as a hex string
func HashString(input string) string {
	hash := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", hash)
}

// HashBytes generates a SHA-256 hash of the input bytes and returns it as a hex string
func HashBytes(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

// GenerateRandomBytes generates random bytes of the specified length
func GenerateRandomBytes(length int) ([]byte, error) {
	data := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return data, nil
}

// GenerateRandomString generates a random string of the specified length
// The string is base64-encoded random bytes, so the actual string length
// may be slightly longer than the requested byte length
func GenerateRandomString(byteLength int) (string, error) {
	data, err := GenerateRandomBytes(byteLength)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// GenerateAPIKey generates a random API key with the specified prefix
// The key after the prefix is 32 random bytes, base64-encoded
func GenerateAPIKey(prefix string) (string, error) {
	randomBytes, err := GenerateRandomBytes(32)
	if err != nil {
		return "", err
	}

	// Use URL-safe base64 encoding without padding
	randomPart := base64.RawURLEncoding.EncodeToString(randomBytes)
	return prefix + randomPart, nil
}
