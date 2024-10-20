package server

import (
	"path/filepath"
	"sort"
	"time"

	"os"

	"github.com/ananthakumaran/paisa/internal/config"
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

func GetSheets(db *gorm.DB) gin.H {
	dir := config.GetSheetDir()
	paths, _ := doublestar.FilepathGlob(dir + "/**/*" + EXTENSION)

	files := []*SheetFile{}
	for _, path := range paths {
		files = append(files, readSheetFileWithVersions(dir, path))
	}

	postings := query.Init(db).All()
	postings = service.PopulateMarketPrice(db, postings)

	return gin.H{"files": files, "postings": postings}
}

func GetSheet(file SheetFile) gin.H {
	dir := config.GetSheetDir()
	return gin.H{"file": readSheetFile(dir, filepath.Join(dir, file.Name))}
}

func DeleteSheetBackups(file SheetFile) gin.H {
	dir := config.GetSheetDir()

	if !config.GetConfig().Readonly {
		versions, _ := filepath.Glob(filepath.Join(dir, file.Name+".backup.*"))
		for _, version := range versions {
			err := os.Remove(version)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	return gin.H{"file": readSheetFileWithVersions(dir, filepath.Join(dir, file.Name))}
}

func SaveSheetFile(db *gorm.DB, file SheetFile) gin.H {
	dir := config.GetSheetDir()

	filePath := filepath.Join(dir, file.Name)
	filePath, err := utils.BuildSubPath(dir, file.Name)
	if err != nil {
		log.Warn(err)
		return gin.H{"saved": false, "message": "Invalid file name"}
	}

	backupPath := filePath + ".backup." + time.Now().Format("2006-01-02-15-04-05.000")

	err = os.MkdirAll(filepath.Dir(filePath), 0700)
	if err != nil {
		log.Warn(err)
		return gin.H{"saved": false, "message": "Failed to create directory"}
	}

	fileStat, err := os.Stat(filePath)
	if err != nil && file.Operation != "overwrite" && file.Operation != "create" {
		log.Warn(err)
		return gin.H{"saved": false, "message": "File does not exist"}
	}

	var perm os.FileMode = 0644
	if err == nil {
		if file.Operation == "create" {
			return gin.H{"saved": false, "message": "File already exists"}
		}

		perm = fileStat.Mode().Perm()
		existingContent, err := os.ReadFile(filePath)
		if err != nil {
			log.Warn(err)
			return gin.H{"saved": false, "message": "Failed to read file"}
		}

		err = os.WriteFile(backupPath, existingContent, perm)
		if err != nil {
			log.Warn(err)
			return gin.H{"saved": false, "message": "Failed to create backup"}
		}
	}

	err = os.WriteFile(filePath, []byte(file.Content), perm)
	if err != nil {
		log.Warn(err)
		return gin.H{"saved": false, "message": "Failed to write file"}
	}

	return gin.H{"saved": true, "file": readSheetFileWithVersions(dir, filePath)}
}

func readSheetFile(dir string, path string) *SheetFile {
	content, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	name, err := filepath.Rel(dir, path)

	return &SheetFile{
		Name:    name,
		Content: string(content),
	}
}

func readSheetFileWithVersions(dir string, path string) *SheetFile {
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

	return &SheetFile{
		Name:     name,
		Content:  string(content),
		Versions: versionPaths,
	}
}
