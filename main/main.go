package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	"net/http"
	"wios_server/conf"
	"wios_server/handlers"
	"wios_server/js"
	"wios_server/middlewares"
)

func main() {
	setLog()
	app := gin.Default()
	middlewares.UseJsoniter(app)
	defer conf.Db.Close()
	defer conf.Rdb.Close()
	startDHTNode()
	handlers.RegisterHandlers(app)
	startServer(app)
}

func setLog() {
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
	// set logrus to default log writer
	gin.DefaultWriter = logrus.StandardLogger().Writer()
	logLevel, err := logrus.ParseLevel(conf.Cfg.LogLevel)
	if err != nil {
		logLevel = logrus.DebugLevel
	}
	logrus.SetLevel(logLevel)
}

type blankValidator struct{}

func (blankValidator) Validate(_ string, _ []byte) error        { return nil }
func (blankValidator) Select(_ string, _ [][]byte) (int, error) { return 0, nil }

func startDHTNode() {
	ctx := context.Background()
	pd, err := base64.StdEncoding.DecodeString(conf.Cfg.DHTConf.PrivateKey)
	if err != nil {
		panic(err)
	}
	pKey, err := crypto.UnmarshalEd25519PrivateKey(pd)
	if err != nil {
		panic(err)
	}
	host, err := libp2p.New(
		libp2p.ListenAddrStrings(conf.Cfg.DHTConf.ListenAddrs...),
		libp2p.Identity(pKey),
	)
	if err != nil {
		logrus.Fatalf("Failed to create libp2p host: %v", err)
	}
	baseOpts := []dht.Option{
		dht.ProtocolPrefix("/wios"),
		dht.Mode(dht.ModeServer),
		dht.NamespacedValidator("v", blankValidator{}),
	}
	kadDHT, err := dht.New(ctx, host, baseOpts...)
	if err != nil {
		logrus.Fatalf("Failed to create DHT: %v", err)
	}
	if err := bootstrapDHT(ctx, kadDHT, conf.Cfg.DHTConf.BootstrapNodes...); err != nil {
		logrus.Fatalf("Failed to bootstrap DHT: %v", err)
	}
	for _, addr := range host.Addrs() {
		logrus.Infof("DHT node on: %s/p2p", addr)
	}
	logrus.Infof("DHT node[%s] started successfully", host.ID())
	js.RegExport("dht", kadDHT)
}
func bootstrapDHT(ctx context.Context, kadDHT *dht.IpfsDHT, bootstrapPeers ...string) error {
	// Parse the bootstrap peer addresses
	for _, peerAddr := range bootstrapPeers {
		addr, err := multiaddr.NewMultiaddr(peerAddr)
		if err != nil {
			logrus.Warningf("failed to parse multiaddr %s: %w", peerAddr, err)
			continue
		}
		info, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			logrus.Warningf("failed to get AddrInfo from multiaddr %s: %w", peerAddr, err)
			continue
		}
		// Connect to the bootstrap peer
		if err := kadDHT.Host().Connect(ctx, *info); err != nil {
			logrus.Warningf("failed to connect to bootstrap peer %s: %w", peerAddr, err)
			continue
		}
	}
	// Bootstrap the DHT
	return kadDHT.Bootstrap(ctx)
}

func startServer(app *gin.Engine) {
	var err error
	if conf.Cfg.EnableHttps {
		err = http.ListenAndServeTLS(":"+conf.Cfg.Port,
			conf.Cfg.CertFile, conf.Cfg.KeyFile, app)
	} else {
		err = app.Run(":" + conf.Cfg.Port)
	}
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
