package js

import (
	"context"
	pebbleds "github.com/ipfs/go-ds-pebble"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"time"
	"wios_server/conf"
	"wios_server/utils"
)

type blankValidator struct{}

func (blankValidator) Validate(_ string, _ []byte) error        { return nil }
func (blankValidator) Select(_ string, _ [][]byte) (int, error) { return 0, nil }

func newDHTNode(log *zap.Logger, pKey crypto.PrivKey, config *conf.Config, ds *pebbleds.Datastore) *dht.IpfsDHT {
	ctx := context.Background()
	host, err := libp2p.New(
		libp2p.ListenAddrStrings(config.DHTConf.ListenAddrs...),
		libp2p.Identity(pKey),
	)
	if err != nil {
		log.Sugar().Fatalf("Failed to create libp2p host: #{err}")
	}
	rt := config.DHTConf.Routing
	baseOpts := []dht.Option{
		dht.ProtocolPrefix("/wios"),
		dht.Mode(dht.ModeServer),
		dht.NamespacedValidator("v", blankValidator{}),
		dht.Datastore(ds),
		dht.MaxRecordAge(utils.ParseDuration(config.DHTConf.MaxRecordAge, 48*time.Hour)),
		dht.RoutingTableLatencyTolerance(utils.ParseDuration(rt.LatencyTolerance, 10*time.Second)),
		dht.RoutingTableRefreshQueryTimeout(utils.ParseDuration(rt.RefreshQueryTimeout, 10*time.Second)),
		dht.RoutingTableRefreshPeriod(utils.ParseDuration(rt.RefreshInterval, 10*time.Minute)),
	}
	if !rt.AutoRefresh {
		baseOpts = append(baseOpts, dht.DisableAutoRefresh())
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

func regExport(kadDHT *dht.IpfsDHT, config *conf.Config) {
	RegExport("dht", kadDHT)
}

var dhtModule = fx.Options(
	fx.Provide(newDHTNode),
	fx.Invoke(regExport),
)
