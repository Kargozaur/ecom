package hasher_test

import (
	"reg/hasher"
	"strings"
	"testing"
)

func TestHash(t *testing.T) {
	h := hasher.NewArgon2Hasher()

	tests := []struct {
		name     string
		password string
	}{
		{"standart password", "SuperSecret123!"},
		{"empty password", ""},
		{"long password", strings.Repeat("a", 256)},
		{"unicode password", "pwd密码🔒"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := h.Hash(tt.password)
			if err != nil {
				t.Fatalf("Hash() returned an empty string: %v", err)
			}
			if encoded == "" {
				t.Fatal("Hash() returned an empty string")
			}

			parts := strings.Split(encoded, "$")
			if len(parts) != 6 {
				t.Fatalf("wrong hash format, parts: %d, want 6 (%s)", len(parts), encoded)
			}
			if parts[1] != "argon2id" {
				t.Errorf("alg = %q, want argon2id", parts[1])
			}
			if !strings.HasPrefix(parts[2], "v=") {
				t.Errorf("expected prefix v=, got %q", parts[2])
			}
			if !strings.HasPrefix(parts[3], "m=") || !strings.Contains(parts[3], "t=") || !strings.Contains(parts[3], "p=") {
				t.Errorf("wrong parameter block: %q", parts[3])
			}
		})
	}
}

func TestCompareHashAndPassword_CorrectPassword(t *testing.T) {
	h := hasher.NewArgon2Hasher()
	password := "correct-horse-battery-staple"

	encoded, err := h.Hash(password)
	if err != nil {
		t.Fatalf("Hash() error: %v", err)
	}

	ok, err := h.CompareHashAndPassword(password, encoded)
	if err != nil {
		t.Fatalf("CompareHashAndPassword() returned an error: %v", err)
	}
	if !ok {
		t.Error("CompareHashAndPassword() = false, want true")
	}
}

func TestCompareHashAndPassword_WrongPassword(t *testing.T) {
	h := hasher.NewArgon2Hasher()
	encoded, err := h.Hash("real-password")
	if err != nil {
		t.Fatalf("Hash() error: %v", err)
	}

	ok, err := h.CompareHashAndPassword("wrong-password", encoded)
	if err != nil {
		t.Fatalf("CompareHashAndPassword() unexpecter error: %v", err)
	}
	if ok {
		t.Error("CompareHashAndPassword() = true, want false")
	}
}

func TestCompareHashAndPassword_InvalidHashFormat(t *testing.T) {
	h := hasher.NewArgon2Hasher()

	tests := []struct {
		name       string
		encodedPwd string
	}{
		{"empty string", ""},
		{"random", "not-a-valid-hash"},
		{"not enough parts", "$argon2id$v=19$m=65536,t=3,p=2"},
		{"corrupted base64 salt", "$argon2id$v=19$m=65536,t=3,p=2$!!!not-base64!!!$aGFzaA"},
		{"corrupted base64 hash", "$argon2id$v=19$m=65536,t=3,p=2$c2FsdA$!!!not-base64!!!"},
		{"corrupted version block", "$argon2id$version=19$m=65536,t=3,p=2$c2FsdA$aGFzaA"},
		{"corrupted parameter block", "$argon2id$v=19$broken-params$c2FsdA$aGFzaA"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := h.CompareHashAndPassword("any-password", tt.encodedPwd)
			if err == nil {
				t.Error("error expected, but got nil insted")
			}
			if ok {
				t.Error("CompareHashAndPassword() = true, want false")
			}
		})
	}
}
