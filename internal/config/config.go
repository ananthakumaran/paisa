package config

import "github.com/spf13/viper"

func DefaultCurrency() string {
	if viper.IsSet("default_currency") {
		return viper.GetString("default_currency")
	}
	return "INR"
}
