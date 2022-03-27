package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "paisa",
	Short: "A command line tool to manager personal finance",
	Long:  "",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./paisa.yaml)")
}

func initConfig() {
	log.SetFormatter(&log.TextFormatter{DisableTimestamp: true})

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
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
