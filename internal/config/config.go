package config

import (
	"sync"

	"github.com/spf13/viper"
)

type Config struct {
	DefaultCurrency string `json:"default_currency"`
	Locale          string `json:"locale"`
	JournalPath     string `json:"journal_path"`
}

var configLock sync.Once

var config Config

func loadConfig() {
	if viper.IsSet("default_currency") {
		config.DefaultCurrency = viper.GetString("default_currency")
	} else {
		config.DefaultCurrency = "INR"
	}

	if viper.IsSet("locale") {
		config.Locale = viper.GetString("locale")
	} else {
		config.Locale = "en-IN"
	}

	config.JournalPath = viper.GetString("journal_path")
}

func GetConfig() Config {
	configLock.Do(loadConfig)
	return config
}

func JournalPath() string {
	configLock.Do(loadConfig)
	return config.JournalPath
}

func DefaultCurrency() string {
	configLock.Do(loadConfig)
	return config.DefaultCurrency
}
