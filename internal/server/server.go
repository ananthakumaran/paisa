package server

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func Listen() {
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	router.SetTrustedProxies(nil)
	router.Static("/static", "web/static")
	router.StaticFile("/", "web/static/index.html")
	log.Info("Listening on 8500")
	router.Run(":8500")
}
