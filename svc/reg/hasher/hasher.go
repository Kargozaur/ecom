package hasher

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type IHasher interface {
	Hash(string) (string, error)
	CompareHashAndPassword(string, string) (bool, error)
}

type Argon2Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

type Hasher struct {
	params *Argon2Params
}

func NewArgon2Hasher() *Hasher {
	defaultParams := &Argon2Params{
		Memory:      64 * 1024,
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}
	return &Hasher{params: defaultParams}
}

func (h *Hasher) Hash(password string) (string, error) {
	if len(password) > 128 {
		return "", ErrStringTooLong
	}
	hash, err := h.generateFromPassword(password)
	if err != nil {
		return "", err
	}
	return hash, nil
}

func (h *Hasher) CompareHashAndPassword(pwd, encodedPwd string) (bool, error) {
	if len(pwd) > 128 {
		return false, ErrStringTooLong
	}
	p, salt, hash, err := h.decodeHash(encodedPwd)
	if err != nil {
		return false, err
	}
	other := argon2.IDKey([]byte(pwd), salt, p.Iterations, p.Memory, p.Parallelism, p.KeyLength)
	if subtle.ConstantTimeCompare(hash, other) == 1 {
		return true, nil
	}
	return false, nil
}

func (h *Hasher) generateFromPassword(password string) (string, error) {
	salt, err := h.generateRandomBytes(h.params.SaltLength)
	if err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password),
		salt,
		h.params.Iterations,
		h.params.Memory,
		h.params.Parallelism,
		h.params.KeyLength)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, h.params.Memory, h.params.Iterations,
		h.params.Parallelism,
		b64Salt,
		b64Hash)
	return encodedHash, nil
}

func (h *Hasher) generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, ErrFailedToGenerateBytes
	}
	return b, nil
}

func (h *Hasher) decodeHash(encodedHash string) (*Argon2Params, []byte, []byte, error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, nil, nil, ErrInvalidHash
	}
	var version int
	_, err := fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}
	p := &Argon2Params{}
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.Memory, &p.Iterations, &p.Parallelism)
	if err != nil {
		return nil, nil, nil, err
	}
	salt, err := base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, err
	}
	p.SaltLength = uint32(len(salt))
	hash, err := base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return nil, nil, nil, err
	}
	p.KeyLength = uint32(len(hash))
	return p, salt, hash, nil
}
