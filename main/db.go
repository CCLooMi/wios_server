package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"wios_server/conf"
)

// 初始化数据库连接
func InitDB(config *conf.Config) *sql.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.DB.User, config.DB.Password, config.DB.Host, config.DB.Port, config.DB.Name)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	return db
}
