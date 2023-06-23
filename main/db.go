package main

import (
	"fmt"
	"wios_server/entity"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// 初始化数据库连接
func InitDB(config *Config) *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.DB.User, config.DB.Password, config.DB.Host, config.DB.Port, config.DB.Name)

	db, err := gorm.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	// 自动迁移表结构
	db.AutoMigrate(&entity.User{})

	return db
}
