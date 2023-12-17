package handlers

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"wios_server/middlewares"
)

func RegisterHandlers(app *gin.Engine, db *sql.DB) {
	app.Use(middlewares.ApplyConfig)
	HandleWebSocket(app, db)
	HandleFileUpload(app, db)
	ServerUploadFile(app, db)
	NewUserController(app, db)
	NewMenuController(app, db)
	NewApiController(app, db)
	ServerStaticDir(app)
}
