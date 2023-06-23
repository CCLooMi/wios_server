package handlers

import (
	"github.com/jinzhu/gorm"
	"github.com/kataras/iris/v12"
)

// 注册路由
func RegisterHandlers(app *iris.Application, db *gorm.DB) {
	app.Get("/users/byPage", GetUserListHandler(db))
	app.Get("/upload/{fileId:string}", getFileFromUploadDirHandler(db))
	serverStaticDir(app)
}
