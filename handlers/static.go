package handlers

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func serverStaticDir(app *gin.Engine) {
	//group := app.Group("/wios")
	//group.GET("/index.html",xxxHandler)
	// 映射静态文件目录
	app.Static("/wios", "./static/public/wios")
}

func getFileFromUploadDirHandler(db *sql.DB) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
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
	}

}

func getRealPath(fid string) string {
	// 将 fid 参数转换为字节数组
	bid, err := hex.DecodeString(fid)
	if err != nil {
		return ""
	}

	// 从字节数组中获取 a 和 b 的值
	a := int(bid[0])
	b := int(bid[1])

	// 返回结果
	return fmt.Sprintf("static/upload/%d/%d/%s/0", a, b, fid)
}
