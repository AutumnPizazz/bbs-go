package secureconfig

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"strings"

	"bbs-go/internal/pkg/config"
)

const prefix = "enc:v1:"

func Encrypt(value string) (string, error) {
	if IsEncrypted(value) {
		return value, nil
	}
	block, err := newBlock()
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, []byte(value), nil)
	return prefix + base64.RawStdEncoding.EncodeToString(sealed), nil
}

func Decrypt(value string) (string, error) {
	if !IsEncrypted(value) {
		return value, nil
	}
	block, err := newBlock()
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	data, err := base64.RawStdEncoding.DecodeString(strings.TrimPrefix(value, prefix))
	if err != nil || len(data) < gcm.NonceSize() {
		return "", errors.New("invalid encrypted configuration")
	}
	nonce, sealed := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, sealed, nil)
	if err != nil {
		return "", errors.New("cannot decrypt configuration")
	}
	return string(plain), nil
}

func IsEncrypted(value string) bool { return strings.HasPrefix(value, prefix) }

func newBlock() (cipher.Block, error) {
	key := strings.TrimSpace(os.Getenv("BBSGO_CONFIG_ENCRYPTION_KEY"))
	if key == "" && config.Instance != nil {
		key = strings.TrimSpace(config.Instance.Security.EncryptionKey)
	}
	if key == "" {
		return nil, errors.New("BBSGO_CONFIG_ENCRYPTION_KEY or security.encryptionKey is required")
	}
	digest := sha256.Sum256([]byte(key))
	return aes.NewCipher(digest[:])
}
