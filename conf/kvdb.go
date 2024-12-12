package conf

import (
	"context"
	"github.com/cockroachdb/pebble"
	"github.com/dustin/go-humanize"
	pebbleds "github.com/ipfs/go-ds-pebble"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

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
			go func() {
				log.Info("closing datastore")
				if err := ds.Close(); err != nil {
					log.Sugar().Errorf("failed to close datastore: %w", err)
				}
			}()
			return nil
		},
	})
	return ds, err
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
