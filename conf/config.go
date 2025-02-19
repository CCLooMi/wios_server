package conf

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"github.com/CCLooMi/sql-mak/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"go.uber.org/fx"
	"go.uber.org/zap"
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
type FutuApiConfig struct {
	Enable  bool   `yaml:"enable"`
	ApiAddr string `yaml:"api_addr"`
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
	DisableServer bool                `yaml:"disable_server"`
	EnableHttps   bool                `yaml:"enable_https"`
	CertFile      string              `yaml:"https_cert_file"`
	KeyFile       string              `yaml:"https_key_file"`
	HostConf      map[string]HostConf `yaml:"host_conf"`
	DHTConf       DHTConfig           `yaml:"dht"`
	DatastoreConf DatastoreConfig     `yaml:"datastore"`
	SysConf       map[string]interface{}
	FutuApiConf   FutuApiConfig `yaml:"futu_api_conf"`
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

var Module = fx.Options(
	fx.Provide(
		loadConfig,
		setLog,
		initDB,
		initPebble,
		initPeerId,
	),
	//use the zap logger for fx
	//fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
	//	return &fxevent.ZapLogger{Logger: log}
	//}),
	fx.Invoke(LoadSysCfg),
)
