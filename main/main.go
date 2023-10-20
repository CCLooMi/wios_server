package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"wios_server/conf"
	"wios_server/handlers"
)

func main() {
	app := gin.Default()
	// 初始化数据库
	db := InitDB(conf.Cfg)
	defer db.Close()
	// 注册路由
	handlers.RegisterHandlers(app, db)
	// 启动HTTP服务
	// app.Run(":8080")
	app.Run(fmt.Sprintf(":%d", conf.Cfg.Port))
}
