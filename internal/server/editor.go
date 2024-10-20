package server

import (
	"path/filepath"
	"sort"
	"time"

	"os"

	"github.com/ananthakumaran/paisa/internal/config"
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

func GetFiles(db *gorm.DB) gin.H {
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
		files = append(files, readLedgerFileWithVersions(dir, path))
	}

	return gin.H{"files": files, "accounts": accounts, "payees": payees, "commodities": commodities}
}

func GetFile(file LedgerFile) gin.H {
	path := config.GetJournalPath()
	dir := filepath.Dir(path)
	return gin.H{"file": readLedgerFile(dir, filepath.Join(dir, file.Name))}
}

func DeleteBackups(file LedgerFile) gin.H {
	path := config.GetJournalPath()
	dir := filepath.Dir(path)

	if !config.GetConfig().Readonly {
		versions, _ := filepath.Glob(filepath.Join(dir, file.Name+".backup.*"))
		for _, version := range versions {
			err := os.Remove(version)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	return gin.H{"file": readLedgerFileWithVersions(dir, filepath.Join(dir, file.Name))}
}

func SaveFile(db *gorm.DB, file LedgerFile) gin.H {
	errors, _, err := validateFile(file)
	if err != nil {
		return gin.H{"errors": errors, "saved": false, "message": "Validation failed"}
	}

	path := config.GetJournalPath()
	dir := filepath.Dir(path)

	filePath, err := utils.BuildSubPath(dir, file.Name)
	if err != nil {
		log.Warn(err)
		return gin.H{"errors": errors, "saved": false, "message": "Invalid file name"}
	}

	backupPath := filePath + ".backup." + time.Now().Format("2006-01-02-15-04-05.000")

	err = os.MkdirAll(filepath.Dir(filePath), 0700)
	if err != nil {
		log.Warn(err)
		return gin.H{"errors": errors, "saved": false, "message": "Failed to create directory"}
	}

	fileStat, err := os.Stat(filePath)
	if err != nil && file.Operation != "overwrite" && file.Operation != "create" {
		log.Warn(err)
		return gin.H{"errors": errors, "saved": false, "message": "File does not exist"}
	}

	var perm os.FileMode = 0644
	if err == nil {
		if file.Operation == "create" {
			return gin.H{"errors": errors, "saved": false, "message": "File already exists"}
		}

		perm = fileStat.Mode().Perm()
		existingContent, err := os.ReadFile(filePath)
		if err != nil {
			log.Warn(err)
			return gin.H{"errors": errors, "saved": false, "message": "Failed to read file"}
		}

		err = os.WriteFile(backupPath, existingContent, perm)
		if err != nil {
			log.Warn(err)
			return gin.H{"errors": errors, "saved": false, "message": "Failed to create backup"}
		}
	}

	err = os.WriteFile(filePath, []byte(file.Content), perm)
	if err != nil {
		log.Warn(err)
		return gin.H{"errors": errors, "saved": false, "message": "Failed to write file"}
	}

	Sync(db, SyncRequest{Journal: true})

	return gin.H{"errors": errors, "saved": true, "file": readLedgerFileWithVersions(dir, filePath)}
}

func ValidateFile(file LedgerFile) gin.H {
	errors, output, _ := validateFile(file)
	return gin.H{"errors": errors, "output": output}
}

func validateFile(file LedgerFile) ([]ledger.LedgerFileError, string, error) {
	path := config.GetJournalPath()

	tmpfile, err := os.CreateTemp(filepath.Dir(path), "paisa-tmp-")
	if err != nil {
		log.Fatal(err)
	}

	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(file.Content)); err != nil {
		log.Fatal(err)
	}

	if err := tmpfile.Close(); err != nil {
		log.Fatal(err)
	}

	return ledger.Cli().ValidateFile(tmpfile.Name())
}

func readLedgerFile(dir string, path string) *LedgerFile {
	content, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	name, err := filepath.Rel(dir, path)

	return &LedgerFile{
		Name:    name,
		Content: string(content),
	}
}

func readLedgerFileWithVersions(dir string, path string) *LedgerFile {
	content, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	versions, _ := filepath.Glob(filepath.Join(filepath.Dir(path), filepath.Base(path)+".backup.*"))
	versionPaths := lo.Map(versions, func(path string, _ int) string {
		name, err := filepath.Rel(dir, path)
		if err != nil {
			log.Fatal(err)
		}

		return name
	})
	sort.Sort(sort.Reverse(sort.StringSlice(versionPaths)))

	name, err := filepath.Rel(dir, path)
	if err != nil {
		log.Fatal(err)
	}

	return &LedgerFile{
		Name:     name,
		Content:  string(content),
		Versions: versionPaths,
	}
}
