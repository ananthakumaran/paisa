package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/ananthakumaran/paisa/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	viper.SetConfigName("paisa.yaml")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}

	db, err := gorm.Open(sqlite.Open(viper.GetString("db_path")), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	model.Sync(db)
}
