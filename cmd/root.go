package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configFile string

var rootCmd = &cobra.Command{
	Use:   "paisa",
	Short: "A command line tool to manager personal finance",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is ./paisa.yaml)")
}

func initConfig() {
	log.SetFormatter(&log.TextFormatter{DisableTimestamp: true, ForceColors: true})

	currentCommand, _, _ := rootCmd.Find(os.Args[1:])
	if currentCommand.Name() == "init" {
		return
	}

	if envConfigFile := os.Getenv("PAISA_CONFIG"); envConfigFile != "" {
		viper.SetConfigFile(envConfigFile)
	} else if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName("paisa.yaml")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
	}

	err := viper.ReadInConfig()
	if err == nil {
		log.Info("Using config file: ", viper.ConfigFileUsed())
	} else {
		log.Fatal(err)
	}
}
