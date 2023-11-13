package conf

import (
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
)

// FileServer config
type FileServerConfig struct {
	SaveDir string `yaml:"saveDir"`
	Path    string `yaml:"path"`
	MaxSize int64  `yaml:"maxSize"`
}

// DBConfig 包含 MySQL 数据库连接信息
type DBConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

// Config 包含应用程序的所有配置信息
type Config struct {
	FileServer FileServerConfig `yaml:"fileServer"`
	DB         DBConfig         `yaml:"db"`
	LogLevel   string           `yaml:"log_level"`
	Port       int              `yaml:"port"`
}

func LoadConfig(configFile string) (*Config, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}
	}
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

var Cfg *Config

func init() {
	cfgName := "config.yaml"
	config, err := LoadConfig("conf/" + cfgName)
	if err != nil {
		config, err = LoadConfig(cfgName)
		if err != nil {
			logrus.Warnf("failed to load config file: %v", err)
			//panic(err)
		}
	}
	Cfg = config
}
