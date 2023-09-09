package binary

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	log "github.com/sirupsen/logrus"
)

var cachedLedgerBinaryPath string

func LedgerBinaryPath() string {
	if cachedLedgerBinaryPath != "" {
		return cachedLedgerBinaryPath
	}

	path, err := exec.LookPath("ledger")
	if err == nil {
		cachedLedgerBinaryPath = path
		return path
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatal(err)
	}

	binDir := filepath.Join(cacheDir, "paisa")
	binaryPath := "ledger"
	if runtime.GOOS == "windows" {
		binaryPath += ".exe"
	}

	path = filepath.Join(binDir, binaryPath)
	err = stage(path, ledgerBinary, 0750)
	if err != nil {
		log.Fatal(err)
	}

	cachedLedgerBinaryPath = path
	return path
}

//go:embed ledger
var ledgerBinary []byte

// EmbeddedBinaryNeedsUpdate returns true if the provided embedded binary file should
// be updated. This determination is based on the modification times and file sizes of both
// the provided executable and the embedded executable. It is expected that the embedded binary
// modification times should match the main `paisa` executable.
func embeddedBinaryNeedsUpdate(exinfo os.FileInfo, embeddedBinaryPath string, size int64) bool {
	if pathinfo, err := os.Stat(embeddedBinaryPath); err == nil {
		return !exinfo.ModTime().Equal(pathinfo.ModTime()) || pathinfo.Size() != size
	}

	// If the stat fails, the file is either missing or permissions are missing
	// to read this -- let above know that an update should be attempted.

	return true
}

// Stage ...
func stage(p string, binData []byte, filemode os.FileMode) error {
	log.Debugf("Staging '%s'", p)

	err := os.MkdirAll(filepath.Dir(p), filemode)
	if err != nil {
		return fmt.Errorf("failed to create dir '%s': %w", filepath.Dir(p), err)
	}

	selfexe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("unable to determine current executable: %w", err)
	}

	exinfo, err := os.Stat(selfexe)
	if err != nil {
		return fmt.Errorf("unable to stat '%s': %w", selfexe, err)
	}

	if !embeddedBinaryNeedsUpdate(exinfo, p, int64(len(binData))) {
		log.Debug("Re-use existing file:", p)
		return nil
	}

	infile, err := os.Open(selfexe)
	if err != nil {
		return fmt.Errorf("unable to open executable '%s': %w", selfexe, err)
	}
	defer infile.Close()

	log.Debugf("Writing static file: '%s'", p)

	_ = os.Remove(p)
	err = os.WriteFile(p, binData, 0550)
	if err != nil {
		return fmt.Errorf("unable to copy to '%s': %w", p, err)
	}

	// In order to properly determine if an update of an embedded binary file is needed,
	// the staged embedded binary needs to have the same modification time as the `paisa`
	// executable.
	if err := os.Chtimes(p, exinfo.ModTime(), exinfo.ModTime()); err != nil {
		return fmt.Errorf("failed to set file modification times of '%s': %w", p, err)
	}
	return nil
}
