package utils

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/xuri/excelize/v2"
	"net"
	"time"
	"wios_server/conf"
)

// 生成随机ID
func GenerateRandomID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func UUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
func GetFPathByFid(fid string) string {
	bid, err := hex.DecodeString(fid)
	if err != nil {
		return ""
	}
	a := int(bid[0])
	b := int(bid[1])
	return fmt.Sprintf("/%d/%d/%s", a, b, fid)
}

func SaveObjDataToRedis(key string, obj any, expire time.Duration) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	return conf.Rdb.
		Set(conf.Ctx, key, data, expire).
		Err()
}

func GetObjDataFromRedis(key string, out interface{}) error {
	data, err := conf.Rdb.Get(conf.Ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), out)
}

func DelFromRedis(key string) {
	conf.Rdb.Del(conf.Ctx, key)
}
func RandomBytes(len int) []byte {
	b := make([]byte, len)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}
func SHA256(username string, password string, seed []byte) string {
	b := sha256.Sum256(append([]byte(username+password), seed...))
	return hex.EncodeToString(b[:])
}
func LookupDNSRecord(domain, dnsServer, recordType string) ([]string, error) {
	resolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, "udp", dnsServer)
		},
	}
	var records []string
	var err error
	switch recordType {
	case "A", "AAAA", "CNAME", "MX", "NS", "PTR", "SRV", "SOA", "TXT", "CAA", "DS", "DNSKEY":
		switch recordType {
		case "A", "AAAA":
			ips, lookupErr := resolver.LookupIPAddr(context.Background(), domain)
			if lookupErr != nil {
				err = lookupErr
			} else {
				for _, ip := range ips {
					records = append(records, ip.String())
				}
			}
		case "TXT":
			txtRecords, lookupErr := resolver.LookupTXT(context.Background(), domain)
			if lookupErr != nil {
				err = lookupErr
			} else {
				records = txtRecords
			}
		case "CNAME":
			cname, lookupErr := resolver.LookupCNAME(context.Background(), domain)
			if lookupErr != nil {
				err = lookupErr
			} else {
				records = append(records, cname)
			}
		case "MX":
			mxs, lookupErr := resolver.LookupMX(context.Background(), domain)
			if lookupErr != nil {
				err = lookupErr
			} else {
				for _, mx := range mxs {
					records = append(records, mx.Host)
				}
			}
		case "SRV":
			_, srvs, lookupErr := resolver.LookupSRV(context.Background(), "SIP", "TCP", domain)
			if lookupErr != nil {
				err = lookupErr
			} else {
				for _, srv := range srvs {
					records = append(records, srv.Target)
				}
			}
		// Add cases for other record types here
		default:
			err = fmt.Errorf("Unsupported record type: %s", recordType)
		}
	default:
		err = fmt.Errorf("Unsupported record type: %s", recordType)
	}
	return records, err
}

func OpenExcel(path string) (*excelize.File, error) {
	return excelize.OpenFile(path)
}
