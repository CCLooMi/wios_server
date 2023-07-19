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
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}
		c.Next()
	})
	app.GET("/ws", HandleWebSocket(db))
	app.GET("/users/byPage", GetUserListHandler(db))
	app.GET("/upload/{fileId:string}", getFileFromUploadDirHandler(db))
	serverStaticDir(app)
}
