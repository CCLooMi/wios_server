package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"net/http"
	"wios_server/conf"
	"wios_server/handlers"
	"wios_server/js"
	"wios_server/middlewares"
	"wios_server/utils"
)

func main() {
	fxa := fx.New(
		conf.Module,
		utils.Module,
		middlewares.Module,
		handlers.Module,
		js.Module,
		fx.Options(
			fx.Provide(
				newGinApp,
				newDHTNode),
			fx.Invoke(
				exportDHT,
				startServer),
		))
	fxa.Run()
}

func newGinApp() (*gin.Engine, error) {
	return gin.Default(), nil
}

type blankValidator struct{}

func (blankValidator) Validate(_ string, _ []byte) error        { return nil }
func (blankValidator) Select(_ string, _ [][]byte) (int, error) { return 0, nil }

func newDHTNode(log *zap.Logger, pKey crypto.PrivKey, config *conf.Config) *dht.IpfsDHT {
	ctx := context.Background()
	host, err := libp2p.New(
		libp2p.ListenAddrStrings(config.DHTConf.ListenAddrs...),
		libp2p.Identity(pKey),
	)
	if err != nil {
		log.Sugar().Fatalf("Failed to create libp2p host: #{err}")
	}
	baseOpts := []dht.Option{
		dht.ProtocolPrefix("/wios"),
		dht.Mode(dht.ModeServer),
		dht.NamespacedValidator("v", blankValidator{}),
	}
	kadDHT, err := dht.New(ctx, host, baseOpts...)
	if err != nil {
		log.Sugar().Fatalf("Failed to create DHT: %v", err)
	}
	if err := bootstrapDHT(log, ctx, kadDHT, config.DHTConf.BootstrapNodes...); err != nil {
		log.Sugar().Fatalf("Failed to bootstrap DHT: %v", err)
	}
	for _, addr := range host.Addrs() {
		log.Sugar().Infof("DHT node on: %s/p2p", addr)
	}
	log.Sugar().Infof("DHT node[%s] started successfully", host.ID())
	return kadDHT
}
func bootstrapDHT(log *zap.Logger, ctx context.Context, kadDHT *dht.IpfsDHT, bootstrapPeers ...string) error {
	// Parse the bootstrap peer addresses
	for _, peerAddr := range bootstrapPeers {
		addr, err := multiaddr.NewMultiaddr(peerAddr)
		if err != nil {
			log.Sugar().Warnf("failed to parse multiaddr %s: %w", peerAddr, err)
			continue
		}
		info, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			log.Sugar().Warnf("failed to get AddrInfo from multiaddr %s: %w", peerAddr, err)
			continue
		}
		// Connect to the bootstrap peer
		if err := kadDHT.Host().Connect(ctx, *info); err != nil {
			log.Sugar().Warnf("failed to connect to bootstrap peer %s: %w", peerAddr, err)
			continue
		}
	}
	// Bootstrap the DHT
	return kadDHT.Bootstrap(ctx)
}

func exportDHT(kadDHT *dht.IpfsDHT) {
	js.RegExport("dht", kadDHT)
}
func startServer(app *gin.Engine, config *conf.Config) {
	var err error
	if config.EnableHttps {
		err = http.ListenAndServeTLS(":"+config.Port,
			config.CertFile, config.KeyFile, app)
	} else {
		err = app.Run(":" + config.Port)
	}
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
