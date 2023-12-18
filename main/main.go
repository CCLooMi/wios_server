package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"wios_server/conf"
	"wios_server/handlers"
)

func main() {
	setLog()
	app := gin.Default()
	defer conf.Db.Close()
	defer conf.Rdb.Close()
	handlers.RegisterHandlers(app)
	startServer(app)
}

func setLog() {
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
	// set logrus to default log writer
	gin.DefaultWriter = logrus.StandardLogger().Writer()
	logLevel, err := logrus.ParseLevel(conf.Cfg.LogLevel)
	if err != nil {
		logLevel = logrus.DebugLevel
	}
	logrus.SetLevel(logLevel)
}

func startServer(app *gin.Engine) {
	var err error
	if conf.Cfg.EnableHttps {
		err = http.ListenAndServeTLS(":"+conf.Cfg.Port,
			conf.Cfg.CertFile, conf.Cfg.KeyFile, app)
	} else {
		err = app.Run(":" + conf.Cfg.Port)
	}
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
