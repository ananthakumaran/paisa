package server

import (
	"net/http"
	"path/filepath"
	"sort"
	"time"

	"os"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/encryption"
	"github.com/ananthakumaran/paisa/internal/query"
	"github.com/ananthakumaran/paisa/internal/service"
	"github.com/ananthakumaran/paisa/internal/utils"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const EXTENSION = ".paisa"

type SheetFile struct {
	Name      string   `json:"name"`
	Content   string   `json:"content"`
	Versions  []string `json:"versions"`
	Operation string   `json:"operation"`
}

func GetSheets(db *gorm.DB) (gin.H, int) {
	dir := config.GetSheetDir()
	paths, _ := doublestar.FilepathGlob(dir + "/**/*" + EXTENSION)

	files := []*SheetFile{}
	for _, path := range paths {
		sf, err := readSheetFileWithVersions(dir, path)
		if err != nil {
			code, h := ledgerReadErrHTTP(err)
			if code != 0 {
				return h, code
			}
			return gin.H{"error": err.Error()}, http.StatusInternalServerError
		}
		files = append(files, sf)
	}

	postings := query.Init(db).All()
	postings = service.PopulateMarketPrice(db, postings)

	return gin.H{"files": files, "postings": postings}, http.StatusOK
}

func GetSheet(file SheetFile) (gin.H, int) {
	dir := config.GetSheetDir()
	sf, err := readSheetFile(dir, filepath.Join(dir, file.Name))
	if err != nil {
		code, h := ledgerReadErrHTTP(err)
		if code != 0 {
			return h, code
		}
		return gin.H{"error": err.Error()}, http.StatusInternalServerError
	}
	return gin.H{"file": sf}, http.StatusOK
}

func DeleteSheetBackups(file SheetFile) (gin.H, int) {
	dir := config.GetSheetDir()

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

	sf, err := readSheetFileWithVersions(dir, filepath.Join(dir, file.Name))
	if err != nil {
		code, h := ledgerReadErrHTTP(err)
		if code != 0 {
			return h, code
		}
		return gin.H{"error": err.Error()}, http.StatusInternalServerError
	}
	return gin.H{"file": sf}, http.StatusOK
}

func SaveSheetFile(db *gorm.DB, file SheetFile) (gin.H, int) {
	dir := config.GetSheetDir()

	filePath, err := utils.BuildSubPath(dir, file.Name)
	if err != nil {
		log.Warn(err)
		return gin.H{"saved": false, "message": "Invalid file name"}, http.StatusOK
	}

	backupPath := filePath + ".backup." + time.Now().Format("2006-01-02-15-04-05.000")

	err = os.MkdirAll(filepath.Dir(filePath), 0700)
	if err != nil {
		log.Warn(err)
		return gin.H{"saved": false, "message": "Failed to create directory"}, http.StatusOK
	}

	fileStat, err := os.Stat(filePath)
	if err != nil && file.Operation != "overwrite" && file.Operation != "create" {
		log.Warn(err)
		return gin.H{"saved": false, "message": "File does not exist"}, http.StatusOK
	}

	var perm os.FileMode = 0644
	if err == nil {
		if file.Operation == "create" {
			return gin.H{"saved": false, "message": "File already exists"}, http.StatusOK
		}

		perm = fileStat.Mode().Perm()
		existingContent, err := os.ReadFile(filePath)
		if err != nil {
			log.Warn(err)
			return gin.H{"saved": false, "message": "Failed to read file"}, http.StatusOK
		}

		err = os.WriteFile(backupPath, existingContent, perm)
		if err != nil {
			log.Warn(err)
			return gin.H{"saved": false, "message": "Failed to create backup"}, http.StatusOK
		}
	}

	err = encryption.WriteFile(filePath, []byte(file.Content), perm, config.IsEncryptionEnabled())
	if err != nil {
		code, h := ledgerReadErrHTTP(err)
		if code != 0 {
			return h, code
		}
		log.Warn(err)
		return gin.H{"saved": false, "message": "Failed to write file"}, http.StatusOK
	}

	sf, err := readSheetFileWithVersions(dir, filePath)
	if err != nil {
		code, h := ledgerReadErrHTTP(err)
		if code != 0 {
			return h, code
		}
		return gin.H{"saved": false, "message": err.Error()}, http.StatusInternalServerError
	}

	return gin.H{"saved": true, "file": sf}, http.StatusOK
}

func readSheetFile(dir string, path string) (*SheetFile, error) {
	content, err := encryption.ReadFile(path)
	if err != nil {
		return nil, err
	}

	name, err := filepath.Rel(dir, path)
	if err != nil {
		return nil, err
	}

	return &SheetFile{
		Name:    name,
		Content: string(content),
	}, nil
}

func readSheetFileWithVersions(dir string, path string) (*SheetFile, error) {
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

	return &SheetFile{
		Name:     name,
		Content:  string(content),
		Versions: versionPaths,
	}, nil
}
