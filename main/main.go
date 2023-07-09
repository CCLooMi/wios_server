package main

import (
	"fmt"

	"wios_server/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	config, err := LoadConfig()
	if err != nil {
		panic(err)
	}

	app := gin.Default()

	// 初始化数据库
	db := InitDB(config)
	defer db.Close()

	// 注册路由
	handlers.RegisterHandlers(app, db)

	// 启动HTTP服务
	// app.Run(":8080")
	app.Run(fmt.Sprintf(":%d", config.Port))
}
