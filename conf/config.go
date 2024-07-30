package conf

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/CCLooMi/sql-mak/mysql"
	"github.com/cockroachdb/pebble"
	"github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	pebbleds "github.com/ipfs/go-ds-pebble"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	PeerId         string           `yaml:"peer_id"`
	PrivateKey     string           `yaml:"private_key"`
	ListenAddrs    []string         `yaml:"listen_addrs"`
	BootstrapNodes []string         `yaml:"bootstrap_nodes"`
	BucketSize     int              `yaml:"bucket_size"`
	MaxRecordAge   string           `yaml:"max_record_age"`
	Routing        DHTRoutingConfig `yaml:"routing"`
}
type DHTRoutingConfig struct {
	LatencyTolerance    string `yaml:"latency_tolerance"`
	RefreshQueryTimeout string `yaml:"refresh_query_timeout"`
	RefreshInterval     string `yaml:"refresh_interval"`
	AutoRefresh         bool   `yaml:"auto_refresh"`
}
type DatastoreConfig struct {
	Path         string `yaml:"path"`
	Compression  string `yaml:"compression"`
	CacheSize    string `yaml:"cache_size"`
	BytesPerSync string `yaml:"bytes_per_sync"`
	MemTableSize string `yaml:"mem_table_size"`
	MaxOpenFiles int    `yaml:"max_open_files"`
}

// Config 包含应用程序的所有配置信息
type Config struct {
	FileServer    FileServerConfig    `yaml:"fileServer"`
	DB            DBConfig            `yaml:"db"`
	EnableCORS    bool                `yaml:"enable_cors"`
	CORSHosts     []string            `yaml:"cors_host_list"`
	Header        map[string]string   `yaml:"header"`
	LogLevel      string              `yaml:"log_level"`
	Port          string              `yaml:"port"`
	EnableHttps   bool                `yaml:"enable_https"`
	CertFile      string              `yaml:"https_cert_file"`
	KeyFile       string              `yaml:"https_key_file"`
	HostConf      map[string]HostConf `yaml:"host_conf"`
	Redis         RedisConfig         `yaml:"redis"`
	DHTConf       DHTConfig           `yaml:"dht"`
	DatastoreConf DatastoreConfig     `yaml:"datastore"`
	SysConf       map[string]interface{}
}

var CorsHostsMap = make(map[string]bool)

func LoadConfig(configFile string) (*Config, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		cfgName = "config.yaml"
		cfg, err := LoadConfig(cfgName)
		if err != nil {
			return nil, err
		}
		return cfg, nil
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
func initPeerId(log *zap.Logger, config *Config) crypto.PrivKey {
	if config.DHTConf.PeerId != "" {
		pd, err := base64.StdEncoding.DecodeString(config.DHTConf.PrivateKey)
		if err == nil {
			pKey, err := crypto.UnmarshalEd25519PrivateKey(pd)
			if err == nil {
				return pKey
			}
		}
	}
	// Generate a new private key
	privKey, pubKey, err := crypto.GenerateKeyPairWithReader(crypto.Ed25519, 2048, rand.Reader)
	if err != nil {
		log.Sugar().Fatalf("Failed to generate key pair: %s", err)
	}
	// Extract the Peer ID from the public key
	peerID, err := peer.IDFromPublicKey(pubKey)
	if err != nil {
		log.Sugar().Fatalf("Failed to get Peer ID from public key: %s", err)
	}
	config.DHTConf.PeerId = peerID.String()
	pd, _ := privKey.Raw()
	config.DHTConf.PrivateKey = base64.StdEncoding.EncodeToString(pd)
	saveConfigToFile(config, log)
	return privKey
}

var cfgName = "conf/config.yaml"

func LoadSysCfg(db *sql.DB, config *Config) {
	delSet := make(map[string]bool)
	if config.SysConf == nil {
		config.SysConf = make(map[string]interface{})
	}
	for key := range config.SysConf {
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
				config.SysConf[name] = value
				continue
			}
			config.SysConf[name] = jo
		}
		for key := range delSet {
			delete(config.SysConf, key)
		}
		return nil
	})
}
func saveConfigToFile(config *Config, log *zap.Logger) {
	data, err := yaml.Marshal(config)
	if err != nil {
		log.Sugar().Errorf("failed to marshal config to YAML: %w", err)
		return
	}
	if err := os.WriteFile(cfgName, data, 0644); err != nil {
		log.Sugar().Errorf("failed to write YAML to file: %w", err)
		return
	}
}
func loadConfig() (*Config, error) {
	return LoadConfig(cfgName)
}
func initDB(lc fx.Lifecycle, config *Config, log *zap.Logger) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.DB.User, config.DB.Password, config.DB.Host, config.DB.Port, config.DB.Name)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			log.Info("closing database connection")
			return db.Close()
		},
	})
	return db, nil
}
func initRedis(lc fx.Lifecycle, config *Config, log *zap.Logger) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Redis.Host, config.Redis.Port),
		Password: config.Redis.Password,
		DB:       0,
	})
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			log.Info("closing redis connection")
			return rdb.Close()
		},
	})
	return rdb, nil
}
func initPebble(lc fx.Lifecycle, config *Config, log *zap.Logger) (*pebbleds.Datastore, error) {
	var c = config.DatastoreConf
	if c.Path == "" {
		c.Path = "datastore"
	}
	var compression pebble.Compression
	switch cm := c.Compression; cm {
	case "zstd", "":
		compression = pebble.ZstdCompression
	case "snappy":
		compression = pebble.SnappyCompression
	case "default":
		compression = pebble.DefaultCompression
	case "none":
		compression = pebble.NoCompression
	}
	cacheSize := parseBytesI64(c.CacheSize, 8*1024*1024)
	bytesPerSync := parseBytesI32(c.BytesPerSync, 512*1024)
	memTableSize := parseBytes(c.MemTableSize, 64*1024*1024)
	if c.MaxOpenFiles <= 0 {
		c.MaxOpenFiles = 1000
	}
	ds, err := pebbleds.NewDatastore(c.Path, &pebble.Options{
		BytesPerSync: bytesPerSync,
		Cache:        pebble.NewCache(cacheSize),
		MemTableSize: memTableSize,
		MaxOpenFiles: c.MaxOpenFiles,
		Levels: []pebble.LevelOptions{
			{Compression: compression},
		},
	})
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			log.Info("closing datastore")
			return ds.Close()
		},
	})
	return ds, err
}

// zapWriter is a custom io.Writer that writes to a zap logger
type zapWriter struct {
	logger *zap.Logger
	lv     zapcore.Level
}

// Write implements the io.Writer interface for zapWriter
func (w zapWriter) Write(p []byte) (n int, err error) {
	pL := len(p)
	if w.lv == zap.DebugLevel {
		w.logger.Debug(string(p))
		return pL, nil
	}
	if w.lv == zap.InfoLevel {
		w.logger.Info(string(p))
		return pL, nil
	}
	if w.lv == zap.WarnLevel {
		w.logger.Warn(string(p))
		return pL, nil
	}
	if w.lv == zap.ErrorLevel {
		w.logger.Error(string(p))
		return pL, nil
	}
	return pL, nil
}
func setLog(config *Config) *zap.Logger {
	zapCfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.DebugLevel),
		Development: false,
		Encoding:    "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	logger, err := zapCfg.Build()
	if err != nil {
		panic(err)
	}
	defer logger.Sync() // Flushes buffer, if any
	// Replace the global logger
	// Replace zap's global logger
	zap.ReplaceGlobals(logger)
	// Redirect stdlib log output to our logger
	zap.RedirectStdLog(logger)
	// Set Gin to use zap's logger
	gin.DefaultWriter = zapWriter{logger: logger, lv: zap.DebugLevel}
	gin.DefaultErrorWriter = zapWriter{logger: logger, lv: zap.ErrorLevel}
	// Set log level from configuration
	logLevel, err := zapcore.ParseLevel(config.LogLevel)
	if err != nil {
		logLevel = zapcore.DebugLevel
	}
	zapCfg.Level.SetLevel(logLevel)
	return logger
}

func parseBytes(s string, df uint64) uint64 {
	v, err := humanize.ParseBytes(s)
	if err != nil {
		return df
	}
	return v
}
func parseBytesI64(s string, df int64) int64 {
	v, err := humanize.ParseBytes(s)
	if err != nil {
		return df
	}
	return int64(v)
}
func parseBytesI32(s string, df int) int {
	v, err := humanize.ParseBytes(s)
	if err != nil {
		return df
	}
	return int(v)
}

var Module = fx.Options(
	fx.Provide(
		loadConfig,
		setLog,
		initDB,
		initRedis,
		initPebble,
		initPeerId,
	),
	//use the zap logger for fx
	//fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
	//	return &fxevent.ZapLogger{Logger: log}
	//}),
	fx.Invoke(LoadSysCfg),
)
