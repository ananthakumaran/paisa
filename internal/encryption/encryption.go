package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/argon2"
)

// File format:
//   Magic header  "PAISA_ENC\x01"  (10 bytes)
//   Argon2id salt                   (16 bytes)
//   AES-GCM nonce                   (12 bytes)
//   Ciphertext + GCM tag            (rest)

const (
	MagicHeader = "PAISA_ENC\x01"
	magicLen    = 10
	saltLen     = 16
	nonceLen    = 12
	keyLen      = 32 // AES-256

	// Argon2id parameters (OWASP recommended)
	argonTime    = 3
	argonMemory  = 64 * 1024 // 64 MB
	argonThreads = 4
)

var (
	ErrNoPassword     = errors.New("encryption password not set: enter it in the app")
	ErrNotEncrypted   = errors.New("file is not encrypted")
	ErrDecryptFailed  = errors.New("decryption failed: wrong password or corrupted file")
	ErrAlreadyEnabled = errors.New("encryption is already enabled")
)

var password string

// SetPassword sets the in-memory encryption password (may be called again after restart or unlock).
func SetPassword(p string) {
	password = p
}

// GetPassword returns the current encryption password.
func GetPassword() string {
	return password
}

// IsPasswordSet returns whether an encryption password has been configured.
func IsPasswordSet() bool {
	return password != ""
}

// deriveKey uses Argon2id to derive a 256-bit key from a password and salt.
func deriveKey(password string, salt []byte) []byte {
	return argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, keyLen)
}

// Encrypt encrypts plaintext bytes with AES-256-GCM using the given password.
func Encrypt(plaintext []byte, password string) ([]byte, error) {
	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	key := deriveKey(password, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, nonceLen)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// Build output: magic + salt + nonce + ciphertext
	result := make([]byte, 0, magicLen+saltLen+nonceLen+len(ciphertext))
	result = append(result, []byte(MagicHeader)...)
	result = append(result, salt...)
	result = append(result, nonce...)
	result = append(result, ciphertext...)

	return result, nil
}

// Decrypt decrypts data encrypted by Encrypt using the given password.
func Decrypt(data []byte, password string) ([]byte, error) {
	minLen := magicLen + saltLen + nonceLen + 1
	if len(data) < minLen {
		return nil, ErrNotEncrypted
	}

	if string(data[:magicLen]) != MagicHeader {
		return nil, ErrNotEncrypted
	}

	salt := data[magicLen : magicLen+saltLen]
	nonce := data[magicLen+saltLen : magicLen+saltLen+nonceLen]
	ciphertext := data[magicLen+saltLen+nonceLen:]

	key := deriveKey(password, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrDecryptFailed
	}

	return plaintext, nil
}

// IsEncrypted checks whether the given data starts with the encryption magic header.
func IsEncrypted(data []byte) bool {
	if len(data) < magicLen {
		return false
	}
	return string(data[:magicLen]) == MagicHeader
}

// IsFileEncrypted checks whether a file on disk is encrypted.
func IsFileEncrypted(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	header := make([]byte, magicLen)
	n, err := f.Read(header)
	if err != nil || n < magicLen {
		return false
	}
	return string(header) == MagicHeader
}

// ReadFile reads a file, transparently decrypting it if encrypted.
// Returns the plaintext content.
func ReadFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if !IsEncrypted(data) {
		return data, nil
	}

	if !IsPasswordSet() {
		return nil, ErrNoPassword
	}

	return Decrypt(data, password)
}

// WriteFile writes content to a file, encrypting it if encryption is enabled.
func WriteFile(path string, content []byte, perm os.FileMode, encryptionEnabled bool) error {
	if encryptionEnabled {
		if !IsPasswordSet() {
			return ErrNoPassword
		}
		encrypted, err := Encrypt(content, password)
		if err != nil {
			return err
		}
		return os.WriteFile(path, encrypted, perm)
	}

	return os.WriteFile(path, content, perm)
}

// DecryptDir decrypts all encrypted files from srcDir into dstDir, preserving
// directory structure. Only files matching the given extension are processed.
// Non-encrypted files are copied as-is.
func DecryptDir(srcDir, dstDir, ext string) error {
	if !IsPasswordSet() {
		return ErrNoPassword
	}

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dstDir, rel)

		if info.IsDir() {
			return os.MkdirAll(dstPath, 0700)
		}

		if ext != "" && !strings.Contains(filepath.Base(path), ext) {
			// Copy non-ledger files as-is (e.g., .dat, include files)
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			return os.WriteFile(dstPath, data, 0600)
		}

		content, err := ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to decrypt %s: %w", path, err)
		}

		return os.WriteFile(dstPath, content, 0600)
	})
}

// EncryptExistingFiles encrypts all plaintext ledger files in a directory.
func EncryptExistingFiles(dir, ext string) (int, error) {
	if !IsPasswordSet() {
		return 0, ErrNoPassword
	}

	count := 0
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if ext != "" && !strings.Contains(filepath.Base(path), ext) {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		if IsEncrypted(data) {
			return nil // already encrypted
		}

		encrypted, err := Encrypt(data, password)
		if err != nil {
			return fmt.Errorf("failed to encrypt %s: %w", path, err)
		}

		err = os.WriteFile(path, encrypted, info.Mode().Perm())
		if err != nil {
			return err
		}

		log.Infof("Encrypted: %s", path)
		count++
		return nil
	})

	return count, err
}

// PrepareDecryptedJournal checks whether any files in the journal directory
// are encrypted. If so, it decrypts them all into a temp directory and returns
// the path to the equivalent journal file inside that temp dir plus a cleanup
// function. If no files are encrypted, it returns the original path and a no-op
// cleanup function.
func PrepareDecryptedJournal(journalPath string) (string, func(), error) {
	dir := filepath.Dir(journalPath)
	base := filepath.Base(journalPath)

	// Quick check: is the main journal file encrypted?
	if !IsFileEncrypted(journalPath) {
		return journalPath, func() {}, nil
	}

	if !IsPasswordSet() {
		return "", func() {}, ErrNoPassword
	}

	tmpDir, err := os.MkdirTemp("", "paisa-dec-*")
	if err != nil {
		return "", func() {}, fmt.Errorf("failed to create temp dir: %w", err)
	}

	cleanup := func() {
		err := os.RemoveAll(tmpDir)
		if err != nil {
			log.Warnf("Failed to clean up temp dir %s: %v", tmpDir, err)
		}
	}

	ext := filepath.Ext(journalPath)
	err = DecryptDir(dir, tmpDir, ext)
	if err != nil {
		cleanup()
		return "", func() {}, fmt.Errorf("failed to decrypt journal files: %w", err)
	}

	return filepath.Join(tmpDir, base), cleanup, nil
}

// DecryptExistingFiles decrypts all encrypted ledger files in a directory back to plaintext.
func DecryptExistingFiles(dir, ext string) (int, error) {
	if !IsPasswordSet() {
		return 0, ErrNoPassword
	}

	count := 0
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if ext != "" && !strings.Contains(filepath.Base(path), ext) {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		if !IsEncrypted(data) {
			return nil // already plaintext
		}

		plaintext, err := Decrypt(data, password)
		if err != nil {
			return fmt.Errorf("failed to decrypt %s: %w", path, err)
		}

		err = os.WriteFile(path, plaintext, info.Mode().Perm())
		if err != nil {
			return err
		}

		log.Infof("Decrypted: %s", path)
		count++
		return nil
	})

	return count, err
}
