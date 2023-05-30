package config

import (
	"sync"

	"github.com/spf13/viper"
)

type Config struct {
	sync.Once
	defaultCurrency string
	journalPath     string
}

var config Config

func loadConfig() {
	if viper.IsSet("default_currency") {
		config.defaultCurrency = viper.GetString("default_currency")
	} else {
		config.defaultCurrency = "INR"
	}
	config.journalPath = viper.GetString("journal_path")
}

func JournalPath() string {
	config.Do(loadConfig)
	return config.journalPath
}

func DefaultCurrency() string {
	config.Do(loadConfig)
	return config.defaultCurrency
}
