package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/samber/lo"
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
	cobra.OnInitialize(Initialize)
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is ./paisa.yaml)")
}

func Initialize() {
	formatter := log.TextFormatter{DisableTimestamp: true, ForceColors: true}
	if os.Getenv("PAISA_DEBUG") == "true" {
		log.SetReportCaller(true)
		log.SetLevel(log.DebugLevel)
		formatter.CallerPrettyfier = func(f *runtime.Frame) (string, string) {
			s := strings.Split(f.Function, ".")
			funcName := s[len(s)-1]
			return funcName, fmt.Sprintf(" [%s:%d]", path.Base(f.File), f.Line)
		}
	}

	log.SetFormatter(&formatter)

	currentCommand, _, _ := rootCmd.Find(os.Args[1:])

	if !lo.Contains([]string{"serve", "update"}, currentCommand.Name()) {
		return
	}

	InitConfig()

}

func InitConfig() {
	if envConfigFile := os.Getenv("PAISA_CONFIG"); envConfigFile != "" {
		readConfigFile(envConfigFile)
	} else if configFile != "" {
		readConfigFile(configFile)
	} else {
		readConfigFile("paisa.yaml")
	}
}

func readConfigFile(path string) {
	path, err := filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		log.Warn("Failed to read config file: ", path)
		log.Fatal(err)
	}

	err = config.LoadConfig(content, path)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Using config file: ", path)
}
