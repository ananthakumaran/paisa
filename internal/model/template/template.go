package template

import (
	"embed"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"path/filepath"
	"strings"
)

//go:embed all:templates/*
var BuiltinTemplates embed.FS

type TemplateType string

const (
	Builtin TemplateType = "builtin"
	Custom  TemplateType = "custom"
)

type Template struct {
	ID           int          `gorm:"primary_key" json:"id"`
	Name         string       `gorm:"uniqueIndex" json:"name"`
	Content      string       `json:"content"`
	TemplateType TemplateType `json:"template_type"`
}

func All(db *gorm.DB) []Template {
	var templates []Template
	db.Find(&templates)

	dirEntries, err := BuiltinTemplates.ReadDir("templates")
	if err != nil {
		log.Fatal(err)
	}
	for i, f := range dirEntries {
		name := f.Name()
		content, err := BuiltinTemplates.ReadFile(fmt.Sprintf("templates/%s", name))
		if err != nil {
			log.Fatal(err)
		}

		name = strings.TrimSuffix(name, filepath.Ext(name))
		template := Template{ID: -i, Name: name, Content: string(content), TemplateType: Builtin}
		templates = append(templates, template)
	}

	return templates
}

func Upsert(db *gorm.DB, name string, content string) Template {
	template := Template{Name: name, Content: content, TemplateType: Custom}
	result := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		UpdateAll: true,
	}).Create(&template)

	if result.Error != nil {
		log.Fatal(result.Error)
	}

	db.First(&template, "name = ?", name)
	return template
}

func Delete(db *gorm.DB, id int) {
	result := db.Delete(&Template{}, id)

	if result.Error != nil {
		log.Fatal(result.Error)
	}
}
