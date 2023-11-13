package handlers

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"wios_server/middlewares"
)

// 注册路由
func RegisterHandlers(app *gin.Engine, db *sql.DB) {
	// 设置跨域请求的配置
	app.Use(middlewares.SetHeaderCors)
	app.Use(middlewares.LoggerRequestInfo)
	HandleWebSocket(app, db)
	HandleFileUpload(app, db)
	ServerUploadFile(app, db)
	NewUserController(app, db)
	NewMenuController(app, db)
	ServerStaticDir(app)
}
