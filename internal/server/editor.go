package server

import (
	"net/http"
	"path/filepath"
	"sort"
	"time"

	"os"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/encryption"
	"github.com/ananthakumaran/paisa/internal/ledger"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type LedgerFile struct {
	Name      string   `json:"name"`
	Content   string   `json:"content"`
	Versions  []string `json:"versions"`
	Operation string   `json:"operation"`
}

func GetFiles(db *gorm.DB) (gin.H, int) {
	var accounts []string
	var payees []string
	var commodities []string
	db.Model(&posting.Posting{}).Distinct().Pluck("Account", &accounts)
	db.Model(&posting.Posting{}).Distinct().Pluck("Payee", &payees)
	db.Model(&posting.Posting{}).Distinct().Pluck("Commodity", &commodities)

	path := config.GetJournalPath()

	files := []*LedgerFile{}
	dir := filepath.Dir(path)
	paths, _ := doublestar.FilepathGlob(dir + "/**/*" + filepath.Ext(path))

	for _, path = range paths {
		lf, err := readLedgerFileWithVersions(dir, path)
		if err != nil {
			code, h := ledgerReadErrHTTP(err)
			if code != 0 {
				return h, code
			}
			return gin.H{"error": err.Error()}, http.StatusInternalServerError
		}
		files = append(files, lf)
	}

	return gin.H{"files": files, "accounts": accounts, "payees": payees, "commodities": commodities}, http.StatusOK
}

func GetFile(file LedgerFile) (gin.H, int) {
	path := config.GetJournalPath()
	dir := filepath.Dir(path)
	lf, err := readLedgerFile(dir, filepath.Join(dir, file.Name))
	if err != nil {
		code, h := ledgerReadErrHTTP(err)
		if code != 0 {
			return h, code
		}
		return gin.H{"error": err.Error()}, http.StatusInternalServerError
	}
	return gin.H{"file": lf}, http.StatusOK
}

func DeleteBackups(file LedgerFile) (gin.H, int) {
	path := config.GetJournalPath()
	dir := filepath.Dir(path)

	if !config.GetConfig().Readonly {
		versions, _ := filepath.Glob(filepath.Join(dir, file.Name+".backup.*"))
		for _, version := range versions {
			err := os.Remove(version)
			if err != nil {
				log.Warn(err)
				return gin.H{"error": err.Error()}, http.StatusInternalServerError
			}
		}
	}

	lf, err := readLedgerFileWithVersions(dir, filepath.Join(dir, file.Name))
	if err != nil {
		code, h := ledgerReadErrHTTP(err)
		if code != 0 {
			return h, code
		}
		return gin.H{"error": err.Error()}, http.StatusInternalServerError
	}
	return gin.H{"file": lf}, http.StatusOK
}

func SaveFile(db *gorm.DB, file LedgerFile) (gin.H, int) {
	errorsOut, _, err := validateFile(file)
	if err != nil {
		code, h := ledgerReadErrHTTP(err)
		if code != 0 {
			return h, code
		}
		return gin.H{"errors": errorsOut, "saved": false, "message": "Validation failed"}, http.StatusOK
	}

	path := config.GetJournalPath()
	dir := filepath.Dir(path)

	filePath, err := utils.BuildSubPath(dir, file.Name)
	if err != nil {
		log.Warn(err)
		return gin.H{"errors": errorsOut, "saved": false, "message": "Invalid file name"}, http.StatusOK
	}

	backupPath := filePath + ".backup." + time.Now().Format("2006-01-02-15-04-05.000")

	err = os.MkdirAll(filepath.Dir(filePath), 0700)
	if err != nil {
		log.Warn(err)
		return gin.H{"errors": errorsOut, "saved": false, "message": "Failed to create directory"}, http.StatusOK
	}

	fileStat, err := os.Stat(filePath)
	if err != nil && file.Operation != "overwrite" && file.Operation != "create" {
		log.Warn(err)
		return gin.H{"errors": errorsOut, "saved": false, "message": "File does not exist"}, http.StatusOK
	}

	var perm os.FileMode = 0644
	if err == nil {
		if file.Operation == "create" {
			return gin.H{"errors": errorsOut, "saved": false, "message": "File already exists"}, http.StatusOK
		}

		perm = fileStat.Mode().Perm()
		existingContent, err := os.ReadFile(filePath)
		if err != nil {
			log.Warn(err)
			return gin.H{"errors": errorsOut, "saved": false, "message": "Failed to read file"}, http.StatusOK
		}

		err = os.WriteFile(backupPath, existingContent, perm)
		if err != nil {
			log.Warn(err)
			return gin.H{"errors": errorsOut, "saved": false, "message": "Failed to create backup"}, http.StatusOK
		}
	}

	err = encryption.WriteFile(filePath, []byte(file.Content), perm, config.IsEncryptionEnabled())
	if err != nil {
		code, h := ledgerReadErrHTTP(err)
		if code != 0 {
			return h, code
		}
		log.Warn(err)
		return gin.H{"errors": errorsOut, "saved": false, "message": "Failed to write file"}, http.StatusOK
	}

	Sync(db, SyncRequest{Journal: true})

	lf, err := readLedgerFileWithVersions(dir, filePath)
	if err != nil {
		code, h := ledgerReadErrHTTP(err)
		if code != 0 {
			return h, code
		}
		return gin.H{"errors": errorsOut, "saved": false, "message": err.Error()}, http.StatusInternalServerError
	}

	return gin.H{"errors": errorsOut, "saved": true, "file": lf}, http.StatusOK
}

func ValidateFile(file LedgerFile) (gin.H, int) {
	errorsOut, output, err := validateFile(file)
	if err != nil {
		code, h := ledgerReadErrHTTP(err)
		if code != 0 {
			return h, code
		}
		return gin.H{"errors": errorsOut, "output": output}, http.StatusOK
	}
	return gin.H{"errors": errorsOut, "output": output}, http.StatusOK
}

func validateFile(file LedgerFile) ([]ledger.LedgerFileError, string, error) {
	journalPath := config.GetJournalPath()
	dir := filepath.Dir(journalPath)

	decryptedDir, cleanup, err := encryption.PrepareDecryptedJournal(journalPath)
	if err != nil {
		return nil, "", err
	}
	defer cleanup()

	var validateDir string
	if decryptedDir != journalPath {
		validateDir = filepath.Dir(decryptedDir)
	} else {
		validateDir = dir
	}

	tmpfile, err := os.CreateTemp(validateDir, "paisa-tmp-")
	if err != nil {
		return nil, "", err
	}

	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(file.Content)); err != nil {
		return nil, "", err
	}

	if err := tmpfile.Close(); err != nil {
		return nil, "", err
	}

	return ledger.Cli().ValidateFile(tmpfile.Name())
}

func readLedgerFile(dir string, path string) (*LedgerFile, error) {
	content, err := encryption.ReadFile(path)
	if err != nil {
		return nil, err
	}

	name, err := filepath.Rel(dir, path)
	if err != nil {
		return nil, err
	}

	return &LedgerFile{
		Name:    name,
		Content: string(content),
	}, nil
}

func readLedgerFileWithVersions(dir string, path string) (*LedgerFile, error) {
	content, err := encryption.ReadFile(path)
	if err != nil {
		return nil, err
	}

	versions, _ := filepath.Glob(filepath.Join(filepath.Dir(path), filepath.Base(path)+".backup.*"))
	versionPaths := lo.Map(versions, func(path string, _ int) string {
		name, relErr := filepath.Rel(dir, path)
		if relErr != nil {
			return ""
		}
		return name
	})
	versionPaths = lo.Filter(versionPaths, func(s string, _ int) bool {
		return s != ""
	})
	sort.Sort(sort.Reverse(sort.StringSlice(versionPaths)))

	name, err := filepath.Rel(dir, path)
	if err != nil {
		return nil, err
	}

	return &LedgerFile{
		Name:     name,
		Content:  string(content),
		Versions: versionPaths,
	}, nil
}
