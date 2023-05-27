package server

import (
	"io/ioutil"
	"path/filepath"
	"sort"
	"time"

	"os"

	"github.com/ananthakumaran/paisa/internal/ledger"
	"github.com/ananthakumaran/paisa/internal/model/posting"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type LedgerFile struct {
	Name     string   `json:"name"`
	Content  string   `json:"content"`
	Versions []string `json:"versions"`
}

func GetFiles(db *gorm.DB) gin.H {
	var accounts []string
	var payees []string
	var commodities []string
	db.Model(&posting.Posting{}).Distinct().Pluck("Account", &accounts)
	db.Model(&posting.Posting{}).Distinct().Pluck("Payee", &payees)
	db.Model(&posting.Posting{}).Distinct().Pluck("Commodity", &commodities)

	path := viper.GetString("journal_path")

	files := []*LedgerFile{}
	dir := filepath.Dir(path)
	paths, _ := filepath.Glob(dir + "/*.ledger")

	for _, path = range paths {
		files = append(files, readLedgerFileWithVersions(dir, path))
	}

	return gin.H{"files": files, "accounts": accounts, "payees": payees, "commodities": commodities}
}

func GetFile(file LedgerFile) gin.H {
	path := viper.GetString("journal_path")
	dir := filepath.Dir(path)
	return gin.H{"file": readLedgerFile(filepath.Join(dir, file.Name))}
}

func DeleteBackups(file LedgerFile) gin.H {
	path := viper.GetString("journal_path")
	dir := filepath.Dir(path)

	versions, _ := filepath.Glob(filepath.Join(dir, file.Name+".backup.*"))
	for _, version := range versions {
		err := os.Remove(version)
		if err != nil {
			log.Fatal(err)
		}
	}

	return gin.H{"file": readLedgerFileWithVersions(dir, filepath.Join(dir, file.Name))}
}

func SaveFile(db *gorm.DB, file LedgerFile) gin.H {
	errors, err := validateFile(file)
	if err != nil {
		return gin.H{"errors": errors, "saved": false}
	}

	path := viper.GetString("journal_path")
	dir := filepath.Dir(path)
	filePath := filepath.Join(dir, file.Name)
	backupPath := filepath.Join(dir, file.Name+".backup."+time.Now().Format("2006-01-02-15-04-05.000"))
	fileStat, err := os.Stat(filePath)
	if err != nil {
		log.Warn(err)
		return gin.H{"errors": errors, "saved": false}
	}

	existingContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Warn(err)
		return gin.H{"errors": errors, "saved": false}
	}

	err = ioutil.WriteFile(backupPath, existingContent, fileStat.Mode().Perm())
	if err != nil {
		log.Warn(err)
		return gin.H{"errors": errors, "saved": false}
	}

	err = ioutil.WriteFile(filePath, []byte(file.Content), fileStat.Mode().Perm())
	if err != nil {
		log.Warn(err)
		return gin.H{"errors": errors, "saved": false}
	}

	Sync(db, SyncRequest{Journal: true})

	return gin.H{"errors": errors, "saved": true, "file": readLedgerFileWithVersions(dir, filePath)}
}

func ValidateFile(file LedgerFile) gin.H {
	errors, _ := validateFile(file)
	return gin.H{"errors": errors}
}

func validateFile(file LedgerFile) ([]ledger.LedgerFileError, error) {
	path := viper.GetString("journal_path")

	tmpfile, err := ioutil.TempFile(filepath.Dir(path), "paisa-tmp-")
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

func readLedgerFile(path string) *LedgerFile {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	return &LedgerFile{
		Name:    filepath.Base(path),
		Content: string(content),
	}
}

func readLedgerFileWithVersions(dir string, path string) *LedgerFile {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	versions, _ := filepath.Glob(filepath.Join(dir, filepath.Base(path)+".backup.*"))
	versionPaths := lo.Map(versions, func(path string, _ int) string { return filepath.Base(path) })
	sort.Sort(sort.Reverse(sort.StringSlice(versionPaths)))
	return &LedgerFile{
		Name:     filepath.Base(path),
		Content:  string(content),
		Versions: versionPaths,
	}
}
