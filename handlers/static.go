package handlers

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path"
	"wios_server/conf"
	"wios_server/utils"
)

func ServerStaticDir(app *gin.Engine) {
	//group := app.Group("/wios")
	//group.GET("/index.html",xxxHandler)
	// 映射静态文件目录
	app.Static("/wios", "./static/public/wios")
}

func ServerUploadFile(app *gin.Engine, db *sql.DB) {
	app.GET("/upload/:fileId", func(ctx *gin.Context) {
		// 获取 fileId 参数
		fileId := ctx.Param("fileId")

		// 计算文件路径
		filePath := getRealPath(fileId)
		//打印文件路径日志
		fmt.Println(fmt.Sprintf(`filePath:%s`, filePath))

		// 检查文件是否存在
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			ctx.String(http.StatusNotFound, "File not found")
			return
		}

		// 设置响应头
		ctx.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileId))

		// 返回文件内容
		ctx.File(filePath)
	})
}

func getRealPath(fid string) string {
	return path.Join(
		conf.Cfg.FileServer.SaveDir,
		utils.GetFPathByFid(fid), "0")
}
