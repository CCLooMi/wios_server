package handlers

import (
	"database/sql"
	"github.com/gin-gonic/gin"
)

// 注册路由
func RegisterHandlers(app *gin.Engine, db *sql.DB) {
	// 设置跨域请求的配置
	app.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}
		c.Next()
	})
	HandleWebSocket(app, db)
	HandleFileUpload(app, db)
	ServerUploadFile(app, db)
	NewUserController(app, db)
	NewMenuController(app, db)
	ServerStaticDir(app)
}
