package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// DBConfig 包含 MySQL 数据库连接信息
type DBConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Name     string `json:"name"`
	User     string `json:"user"`
	Password string `json:"password"`
}

// Config 包含应用程序的所有配置信息
type Config struct {
	DB       DBConfig `json:"db"`
	LogLevel string   `json:"log_level"`
	Port     int      `json:"port"`
}

// LoadConfig 从环境变量或配置文件中读取配置
func LoadConfig() (*Config, error) {
	// 尝试从配置文件中读取配置
	data, err := os.ReadFile("conf/config.json")
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	var config Config
	var dbHost, dbPort, dbName, dbUser, dbPassword, logLevel, port string
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	// 如果成功读取到配置文件，则使用配置文件中的值覆盖环境变量中的值
	if config.DB.Host != "" {
		dbHost = config.DB.Host
	}

	if config.DB.Port != 0 {
		dbPort = fmt.Sprintf("%d", config.DB.Port)
	}

	if config.DB.Name != "" {
		dbName = config.DB.Name
	}

	if config.DB.User != "" {
		dbUser = config.DB.User
	}

	if config.DB.Password != "" {
		dbPassword = config.DB.Password
	}

	if config.LogLevel != "" {
		logLevel = config.LogLevel
	}

	if config.Port != 0 {
		port = fmt.Sprintf("%d", config.Port)
	}

	// 构造 Config 对象
	config = Config{
		DB: DBConfig{
			Host:     dbHost,
			Port:     parseInt(dbPort),
			Name:     dbName,
			User:     dbUser,
			Password: dbPassword,
		},
		LogLevel: logLevel,
		Port:     parseInt(port),
	}

	return &config, nil
}

// parseInt 将字符串转换为整数，如果无法转换则返回 0
func parseInt(s string) int {
	n := 0
	_, err := fmt.Sscanf(s, "%d", &n)
	if err != nil {
		return 0
	}
	return n
}
