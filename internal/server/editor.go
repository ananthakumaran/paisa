package server

import (
	"io/ioutil"
	"path/filepath"
	"time"

	"os"

	"github.com/ananthakumaran/paisa/internal/ledger"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type LedgerFile struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

func GetFiles() gin.H {
	path := viper.GetString("journal_path")

	files := []*LedgerFile{}
	dir := filepath.Dir(path)
	paths, _ := filepath.Glob(dir + "/*.ledger")

	for _, path = range paths {
		files = append(files, readLedgerFile(path))
	}

	return gin.H{"files": files}
}

func SaveFile(file LedgerFile) gin.H {
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

	return gin.H{"errors": errors, "saved": true}
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

	return ledger.ValidateFile(tmpfile.Name())
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
