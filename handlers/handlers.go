package handlers

import (
	"github.com/gin-gonic/gin"
	"wios_server/middlewares"
)

func RegisterHandlers(app *gin.Engine) {
	app.Use(middlewares.ApplyConfig, middlewares.AuthCheck)
	HandleWebSocket(app)
	HandleFileUpload(app)
	ServerUploadFile(app)
	NewUserController(app)
	NewMenuController(app)
	NewOrgController(app)
	NewRoleController(app)
	NewUploadController(app)
	NewApiController(app)
	NewConfigController(app)
	NewStoreUserController(app)
	NewWppController(app)
	NewWppEventController(app)
	NewWppStoryController(app)
	CorsRevertServer(app)
	ServerStaticDir(app)
}
