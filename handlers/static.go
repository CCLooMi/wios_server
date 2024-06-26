package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path"
	"wios_server/conf"
	"wios_server/middlewares"
	"wios_server/service"
	"wios_server/utils"
)

func ServerStaticDir(app *gin.Engine) {
	app.Static("/wios", "./static/public/wios")
	app.Static("/test", "./static/public/test")
}

func ServerUploadFile(app *gin.Engine) {
	wppServ := service.NewWppService(conf.Db)
	app.GET("/upload/:fileId", func(ctx *gin.Context) {
		fileId := ctx.Param("fileId")
		filePath := getRealPath(fileId)
		// check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			ctx.String(http.StatusNotFound, "File not found")
			return
		}
		// check if file is wpp
		wppId := wppServ.IsWpp(&fileId)
		if wppId != nil {
			stUser := middlewares.GetStoreUserInfo(ctx)
			if stUser == nil {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"message": "Unauthorized",
				})
				return
			}
			// check if user has own wpp
			if !wppServ.CheckPurchased(wppId, stUser.User.Id) {
				ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"message": "Forbidden",
				})
				return
			}
		}
		ctx.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileId))
		// response file data
		ctx.File(filePath)
	})
}

func getRealPath(fid string) string {
	return path.Join(
		conf.Cfg.FileServer.SaveDir,
		utils.GetFPathByFid(fid), "0")
}
