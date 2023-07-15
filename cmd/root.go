package cmd

import (
	"io/ioutil"
	"os"

	"github.com/ananthakumaran/paisa/internal/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
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
		readConfigFile(envConfigFile)
	} else if configFile != "" {
		readConfigFile(configFile)
	} else {
		readConfigFile("paisa.yaml")
	}
}

func readConfigFile(path string) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Warn("Failed to read config file: ", path)
		log.Fatal(err)
	}

	err = config.LoadConfig(content)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Using config file: ", path)
}
