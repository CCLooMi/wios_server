package conf

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/CCLooMi/sql-mak/mysql"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	"wios_server/entity"
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
	Port     string `yaml:"port"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}
type RedisConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Password string `yaml:"password"`
}
type HostConf struct {
	Header map[string]string `yaml:"header"`
}
type DHTConfig struct {
	PeerId         string   `yaml:"peer_id"`
	PrivateKey     string   `yaml:"private_key"`
	ListenAddrs    []string `yaml:"listen_addrs"`
	BootstrapNodes []string `yaml:"bootstrap_nodes"`
}

// Config 包含应用程序的所有配置信息
type Config struct {
	FileServer  FileServerConfig    `yaml:"fileServer"`
	DB          DBConfig            `yaml:"db"`
	EnableCORS  bool                `yaml:"enable_cors"`
	CORSHosts   []string            `yaml:"cors_host_list"`
	Header      map[string]string   `yaml:"header"`
	LogLevel    string              `yaml:"log_level"`
	Port        string              `yaml:"port"`
	EnableHttps bool                `yaml:"enable_https"`
	CertFile    string              `yaml:"https_cert_file"`
	KeyFile     string              `yaml:"https_key_file"`
	HostConf    map[string]HostConf `yaml:"host_conf"`
	Redis       RedisConfig         `yaml:"redis"`
	DHTConf     DHTConfig           `yaml:"dht"`
}

var CorsHostsMap = make(map[string]bool)

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
	for _, host := range config.CORSHosts {
		CorsHostsMap[host] = true
	}
	return &config, nil
}

var Cfg *Config
var Db *sql.DB
var Rdb *redis.Client
var Ctx = context.Background()
var SysCfg = make(map[string]interface{})
var cfgName string

func init() {
	cfgName = "conf/config.yaml"
	config, err := LoadConfig(cfgName)
	if err != nil {
		cfgName = "config.yaml"
		config, err = LoadConfig(cfgName)
		if err != nil {
			logrus.Warnf("failed to load config file: %v", err)
			return
		}
	}
	Db, Rdb = InitDB(config)
	Cfg = config
	initPeerId()
}
func initPeerId() {
	if Cfg.DHTConf.PeerId != "" {
		return
	}
	// Generate a new private key
	privKey, pubKey, err := crypto.GenerateKeyPairWithReader(crypto.Ed25519, 2048, rand.Reader)
	if err != nil {
		logrus.Fatalf("Failed to generate key pair: %s", err)
	}
	// Extract the Peer ID from the public key
	peerID, err := peer.IDFromPublicKey(pubKey)
	if err != nil {
		logrus.Fatalf("Failed to get Peer ID from public key: %s", err)
	}
	Cfg.DHTConf.PeerId = peerID.String()
	pd, _ := privKey.Raw()
	Cfg.DHTConf.PrivateKey = base64.StdEncoding.EncodeToString(pd)
	saveConfigToFile()
}
func InitDB(config *Config) (*sql.DB, *redis.Client) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.DB.User, config.DB.Password, config.DB.Host, config.DB.Port, config.DB.Name)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	LoadSysCfg(db)
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Redis.Host, config.Redis.Port),
		Password: config.Redis.Password,
		DB:       0,
	})
	return db, rdb
}
func LoadSysCfg(db *sql.DB) {
	delSet := make(map[string]bool)
	for key := range SysCfg {
		delSet[key] = true
	}
	entity := entity.Config{}
	mysql.SELECT("c.name", "c.value").
		FROM(&entity, "c").
		Execute(db).ExtractorResultSet(func(rs *sql.Rows) interface{} {
		for rs.Next() {
			var name, value string
			err := rs.Scan(&name, &value)
			if err != nil {
				continue
			}
			delete(delSet, name)
			var jo interface{}
			err = json.Unmarshal([]byte(value), &jo)
			if err != nil {
				SysCfg[name] = value
				continue
			}
			SysCfg[name] = jo
		}
		for key := range delSet {
			delete(SysCfg, key)
		}
		return nil
	})
}

func saveConfigToFile() {
	data, err := yaml.Marshal(Cfg)
	if err != nil {
		logrus.Errorf("failed to marshal config to YAML: %w", err)
		return
	}
	if err := os.WriteFile(cfgName, data, 0644); err != nil {
		logrus.Errorf("failed to write YAML to file: %w", err)
		return
	}
}
