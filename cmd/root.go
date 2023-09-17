package cmd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/adrg/xdg"
	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/ananthakumaran/paisa/internal/generator"
	"github.com/ananthakumaran/paisa/internal/utils"
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
	InitLogger(false, nil)
	currentCommand, _, _ := rootCmd.Find(os.Args[1:])

	if !lo.Contains([]string{"serve", "update"}, currentCommand.Name()) {
		return
	}

	InitConfig()

}

func InitLogger(desktop bool, hook log.Hook) {
	formatter := log.TextFormatter{
		DisableTimestamp: true,
		ForceColors:      !desktop,
		DisableColors:    desktop,
		PadLevelText:     true,
	}
	if os.Getenv("PAISA_DEBUG") == "true" {
		log.SetReportCaller(true)
		log.SetLevel(log.DebugLevel)
		formatter.CallerPrettyfier = func(f *runtime.Frame) (string, string) {
			s := strings.Split(f.Function, ".")
			funcName := s[len(s)-1]
			return funcName, fmt.Sprintf(" [%s:%d]", path.Base(f.File), f.Line)
		}
	}

	if desktop && os.Getenv("PAISA_DEBUG") != "true" {
		log.SetReportCaller(true)
	}

	if desktop {
		cacheDir, err := os.UserCacheDir()
		if err == nil {
			p := filepath.Join(cacheDir, "paisa", "paisa.log")
			err = os.MkdirAll(filepath.Dir(p), 0750)
			if err == nil {
				file, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
				if err == nil {
					log.SetOutput(file)
				}
			}
		}
	}

	log.SetFormatter(&formatter)

	if hook != nil {
		log.AddHook(hook)
	}
}

func InitConfig() {
	xdgDocumentDir := filepath.Join(xdg.UserDirs.Documents, "paisa")
	xdgDocumentPath := filepath.Join(xdgDocumentDir, "paisa.yaml")
	if envConfigFile := os.Getenv("PAISA_CONFIG"); envConfigFile != "" {
		config.LoadConfigFile(envConfigFile)
	} else if configFile != "" {
		config.LoadConfigFile(configFile)
	} else if utils.FileExists("paisa.yaml") {
		config.LoadConfigFile("paisa.yaml")
	} else if utils.FileExists(xdgDocumentPath) {
		config.LoadConfigFile(xdgDocumentPath)
	} else {
		err := os.MkdirAll(xdgDocumentDir, 0755)
		if err != nil {
			log.Fatal(err)
		}
		generator.MinimalConfig(xdgDocumentDir)
		config.LoadConfigFile(xdgDocumentPath)
	}
}
