package crypto

import (
	"strings"
	"testing"
)

func TestGenerateKeyFromPassphrase(t *testing.T) {
	key := GenerateKeyFromPassphrase("test-passphrase")
	if len(key) != 32 {
		t.Errorf("Expected key length of 32, got %d", len(key))
	}

	// Same passphrase should generate same key
	key2 := GenerateKeyFromPassphrase("test-passphrase")
	if string(key) != string(key2) {
		t.Error("Same passphrase should generate same key")
	}

	// Different passphrase should generate different key
	key3 := GenerateKeyFromPassphrase("different-passphrase")
	if string(key) == string(key3) {
		t.Error("Different passphrase should generate different key")
	}
}

func TestEncryptDecryptString(t *testing.T) {
	key := GenerateKeyFromPassphrase("test-passphrase")

	testCases := []struct {
		name  string
		input string
	}{
		{"simple text", "hello world"},
		{"empty string", ""},
		{"special characters", "!@#$%^&*()"},
		{"unicode", "你好世界"},
		{"long text", strings.Repeat("a", 1000)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Encrypt
			encrypted, err := EncryptString(tc.input, key)
			if err != nil {
				t.Fatalf("EncryptString failed: %v", err)
			}

			// Encrypted should be different from original
			if encrypted == tc.input {
				t.Error("Encrypted string should be different from original")
			}

			// Decrypt
			decrypted, err := DecryptString(encrypted, key)
			if err != nil {
				t.Fatalf("DecryptString failed: %v", err)
			}

			// Decrypted should match original
			if decrypted != tc.input {
				t.Errorf("Decrypted string doesn't match original: got %q, want %q", decrypted, tc.input)
			}
		})
	}
}

func TestDecryptStringWithInvalidData(t *testing.T) {
	key := GenerateKeyFromPassphrase("test-passphrase")

	testCases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty string", "", true},
		{"invalid base64", "not-valid-base64!!!", true},
		{"too short", "YWJj", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := DecryptString(tc.input, key)
			if (err != nil) != tc.wantErr {
				t.Errorf("DecryptString() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestDecryptStringWithWrongKey(t *testing.T) {
	key1 := GenerateKeyFromPassphrase("passphrase1")
	key2 := GenerateKeyFromPassphrase("passphrase2")

	plaintext := "secret message"

	encrypted, err := EncryptString(plaintext, key1)
	if err != nil {
		t.Fatalf("EncryptString failed: %v", err)
	}

	// Try to decrypt with wrong key
	_, err = DecryptString(encrypted, key2)
	if err == nil {
		t.Error("Expected error when decrypting with wrong key")
	}
}

func TestGenerateRandomKey(t *testing.T) {
	key1, _ := GenerateKey()
	key2, _ := GenerateKey()

	if len(key1) != 32 {
		t.Errorf("Expected key length of 32, got %d", len(key1))
	}

	if len(key2) != 32 {
		t.Errorf("Expected key length of 32, got %d", len(key2))
	}

	// Keys should be different
	if string(key1) == string(key2) {
		t.Error("Randomly generated keys should be different")
	}
}

func TestHashString(t *testing.T) {
	testCases := []struct {
		name string
		input string
	}{
		{"simple", "hello"},
		{"empty", ""},
		{"special", "!@#$%"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash1 := HashString(tc.input)
			hash2 := HashString(tc.input)

			// Same input should produce same hash
			if hash1 != hash2 {
				t.Error("Same input should produce same hash")
			}

			// Hash should be different from input
			if hash1 == tc.input && tc.input != "" {
				t.Error("Hash should be different from input")
			}

			// Different input should produce different hash
			hash3 := HashString(tc.input + "-modified")
			if hash1 == hash3 {
				t.Error("Different input should produce different hash")
			}
		})
	}
}
