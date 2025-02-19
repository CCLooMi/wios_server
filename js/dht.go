package js

import (
	"context"
	"github.com/ipfs/go-cid"
	pebbleds "github.com/ipfs/go-ds-pebble"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	kbucket "github.com/libp2p/go-libp2p-kbucket"
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
	RegExport("dht", &JDHT{kadDHT})
}

type JDHT struct {
	kdht *dht.IpfsDHT
}

func (d *JDHT) ListPeerIds() []string {
	t := d.kdht.RoutingTable()
	pids := t.ListPeers()
	ids := make([]string, len(pids))
	for i, p := range pids {
		ids[i] = p.String()
	}
	return ids
}

func (d *JDHT) PeerID() string {
	return d.kdht.PeerID().String()
}

func (d *JDHT) FindPeer(ctx context.Context, peerId string) (peer.AddrInfo, error) {
	id, err := peer.Decode(peerId)
	if err != nil {
		return peer.AddrInfo{}, err
	}
	return d.kdht.FindPeer(ctx, id)
}
func (d *JDHT) FindProviders(ctx context.Context, key string) ([]peer.AddrInfo, error) {
	id, err := cid.Parse(key)
	if err != nil {
		return nil, err
	}
	return d.kdht.FindProviders(ctx, id)
}
func (d *JDHT) Provide(ctx context.Context, key string) error {
	id, err := cid.Decode(key)
	if err != nil {
		return err
	}
	return d.kdht.Provide(ctx, id, true)
}
func (d *JDHT) Put(ctx context.Context, key string, value []byte) error {
	return d.kdht.PutValue(ctx, key, value)
}
func (d *JDHT) Get(ctx context.Context, key string) ([]byte, error) {
	return d.kdht.GetValue(ctx, key)
}
func (d *JDHT) RouteTable() *kbucket.RoutingTable {
	return d.kdht.RoutingTable()
}
func (d *JDHT) Ping(ctx context.Context, peerId string) error {
	id, err := peer.Decode(peerId)
	if err != nil {
		return err
	}
	return d.kdht.Ping(ctx, id)
}
func (d *JDHT) FindLocal(ctx context.Context, peerId string) peer.AddrInfo {
	id, err := peer.Decode(peerId)
	if err != nil {
		return peer.AddrInfo{}
	}
	return d.kdht.FindLocal(ctx, id)
}

var dhtModule = fx.Options(
	fx.Provide(newDHTNode),
	fx.Invoke(regExport),
)
