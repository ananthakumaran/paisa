package encryption

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// DirHasEncryptedFiles reports whether any file under dir whose basename contains nameSubstring is encrypted.
func DirHasEncryptedFiles(dir, nameSubstring string) bool {
	var found bool
	_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if nameSubstring != "" && !strings.Contains(filepath.Base(path), nameSubstring) {
			return nil
		}
		if IsFileEncrypted(path) {
			found = true
			return fs.SkipAll
		}
		return nil
	})
	return found
}

// VerifyPasswordForDir walks dir and decrypts every file whose basename contains nameSubstring and is encrypted.
func VerifyPasswordForDir(dir, nameSubstring, passwd string) error {
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if nameSubstring != "" && !strings.Contains(filepath.Base(path), nameSubstring) {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !IsEncrypted(b) {
			return nil
		}
		_, err = Decrypt(b, passwd)
		return err
	})
}
