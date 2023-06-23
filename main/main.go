package main

import (
	"fmt"

	"github.com/kataras/iris/v12"

	"wios_server/handlers"
)

func main() {
	config, err := LoadConfig()
	if err != nil {
		panic(err)
	}

	app := iris.New()

	// 初始化数据库
	db := InitDB(config)
	defer db.Close()

	// 注册路由
	handlers.RegisterHandlers(app, db)

	// 启动HTTP服务
	// app.Run(iris.Addr(":8080"))
	app.Run(iris.Addr(fmt.Sprintf(":%d", config.Port)))
}
