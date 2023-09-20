package server

import (
	"encoding/json"
	"os"

	"github.com/ananthakumaran/paisa/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/icza/backscanner"
	log "github.com/sirupsen/logrus"
)

func GetLogs() gin.H {
	logs := make([]interface{}, 0)
	path, err := config.EnsureLogFilePath()
	if err != nil {
		log.Warn(err)
		return gin.H{"logs": logs}
	}

	file, err := os.Open(path)
	if err != nil {
		return gin.H{"logs": logs}

	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return gin.H{"logs": logs}
	}

	scanner := backscanner.New(file, int(stat.Size()))
	count := 0
	for {
		count++
		line, _, err := scanner.LineBytes()
		if err != nil {
			break
		}
		if count > 10000 {
			break
		}
		var log interface{}
		err = json.Unmarshal(line, &log)
		if err == nil {
			logs = append(logs, log)
		}

	}

	return gin.H{"logs": logs}
}
