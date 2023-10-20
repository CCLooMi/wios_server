package conf

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"os"
)

// FileServer config
type FileServerConfig struct {
	SaveDir string `json:"saveDir"`
}

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
	FileServer FileServerConfig `json:"fileServer"`
	DB         DBConfig         `json:"db"`
	LogLevel   string           `json:"log_level"`
	Port       int              `json:"port"`
}

func LoadConfig(configFile string) (*Config, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

var Cfg *Config

func init() {
	config, err := LoadConfig("conf/config.json")
	if err != nil {
		config, err = LoadConfig("config.json")
		if err != nil {
			logrus.Warnf("failed to load config file: %v", err)
			//panic(err)
		}
	}
	Cfg = config
}
