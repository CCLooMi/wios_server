package handlers

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

// 注册路由
func RegisterHandlers(app *gin.Engine, db *sql.DB) {
	app.GET("/users/byPage", GetUserListHandler(db))
	app.GET("/upload/{fileId:string}", getFileFromUploadDirHandler(db))
	serverStaticDir(app)
}
