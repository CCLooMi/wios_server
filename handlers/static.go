package handlers

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
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
		w := ctx.Query("w")
		h := ctx.Query("h")
		var width, height = 0, 0
		var err error
		if w != "" {
			width, err = strconv.Atoi(w)
			if err != nil {
				width = 0
			}
		}
		if h != "" {
			height, err = strconv.Atoi(h)
			if err != nil {
				height = 0
			}
		}
		// If width or height is provided, resize the image
		if width > 0 || height > 0 {
			absPath0, err := filepath.Abs(filePath)
			if err != nil {
				ctx.String(http.StatusInternalServerError, err.Error())
				return
			}
			fname := fmt.Sprintf("%d_%d.jpg", width, height)
			resizedFilePath := path.Join(filepath.Dir(absPath0), fname)
			if name == "" {
				ctx.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fname))
			}
			_, err = os.Stat(resizedFilePath)
			if err == nil {
				ctx.File(resizedFilePath)
				return
			}
			args := []string{
				absPath0,
				"-s", fmt.Sprintf("%dx%d", width, height),
				"-o", resizedFilePath,
			}
			// Run the vips command to resize the image
			cmd := exec.Command(getVipsThumbnailPath(), args...)
			err = cmd.Run()
			if err != nil {
				fmt.Println(err.Error())
				ctx.String(http.StatusInternalServerError, "Error resizing image")
				return
			}
			// After resizing, send the resized image to the client
			ctx.File(resizedFilePath)
			return
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
func getVipsThumbnailPath() string {
	// check default path
	if path, err := exec.LookPath("vipsthumbnail"); err == nil {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "vipsthumbnail"
	}
	basePath := path.Join(home, "/scoop/apps/libvips/current/bin")
	vipsthumbnailPath := filepath.Join(basePath, "vipsthumbnail")
	_, err = os.Stat(vipsthumbnailPath)
	if err != nil {
		return "vipsthumbnail"
	}
	return vipsthumbnailPath
}
func getRealPath(workDir string, fid string) string {
	return path.Join(
		workDir,
		utils.GetFPathByFid(fid), "0")
}
