package server

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/encryption"
)

func encryptionNeedsUnlock() bool {
	if encryption.IsPasswordSet() {
		return false
	}
	journalPath := config.GetJournalPath()
	journalDir := filepath.Dir(journalPath)
	journalExt := filepath.Ext(journalPath)
	if encryption.DirHasEncryptedFiles(journalDir, journalExt) {
		return true
	}
	sheetDir := config.ResolveSheetDirectory()
	if st, err := os.Stat(sheetDir); err == nil && st.IsDir() {
		if encryption.DirHasEncryptedFiles(sheetDir, ".paisa") {
			return true
		}
	}
	return false
}

func trySetEncryptionPasswordFromInput(raw string) error {
	pass := strings.TrimSpace(raw)
	if pass == "" {
		return errors.New("password is empty")
	}
	journalPath := config.GetJournalPath()
	journalDir := filepath.Dir(journalPath)
	journalExt := filepath.Ext(journalPath)
	if err := encryption.VerifyPasswordForDir(journalDir, journalExt, pass); err != nil {
		return err
	}
	sheetDir := config.ResolveSheetDirectory()
	if st, err := os.Stat(sheetDir); err == nil && st.IsDir() {
		if err := encryption.VerifyPasswordForDir(sheetDir, ".paisa", pass); err != nil {
			return err
		}
	}
	encryption.SetPassword(pass)
	return nil
}
