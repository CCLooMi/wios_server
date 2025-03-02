package handlers

import (
	"database/sql"
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

func ServerUploadFile(app *gin.Engine, db *sql.DB, config *conf.Config, ut *utils.Utils) {
	wppServ := service.NewWppService(db)
	app.GET("/upload/:fileId", func(ctx *gin.Context) {
		fileId := ctx.Param("fileId")
		filePath := getRealPath(config.FileServer.SaveDir, fileId)
		// check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			ctx.String(http.StatusNotFound, "File not found")
			return
		}
		// check if file is wpp
		wppId := wppServ.IsWpp(&fileId)
		if wppId != nil {
			stUser := middlewares.GetStoreUserInfo(ctx, ut)
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
		name := ctx.Query("name")
		if name != "" {
			ctx.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, name))
		} else {
			ctx.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileId))
		}
		// response file data
		ctx.File(filePath)
	})
	app.GET("/upload/:fileId/*fileName", func(ctx *gin.Context) {
		fileId := ctx.Param("fileId")
		fileName := ctx.Param("fileName")
		filePath := path.Join(config.FileServer.SaveDir, utils.GetFPathByFid(fileId), fileName)
		// check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			ctx.String(http.StatusNotFound, "File not found")
			return
		}
		// response file data
		ctx.File(filePath)
	})
}

func getRealPath(workDir string, fid string) string {
	return path.Join(
		workDir,
		utils.GetFPathByFid(fid), "0")
}
