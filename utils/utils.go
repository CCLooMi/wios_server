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
	"os"
	"path"
	"path/filepath"
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
func OpenExcelByFid(fid string) (*excelize.File, error) {
	basePath := path.Join(conf.Cfg.FileServer.SaveDir, GetFPathByFid(fid))
	path := filepath.Join(basePath, "0")
	_, err := os.Stat(path)
	if err == nil {
		return excelize.OpenFile(path)
	}
	return nil, err
}
func SetExcelSheetRow(f *excelize.File, sheet string, cell string, data ...interface{}) error {
	if err := f.SetSheetRow(sheet, cell, &data); err != nil {
		return err
	}
	return nil
}
func DelFileByFid(fid string) bool {
	bid, err := hex.DecodeString(fid)
	if err != nil {
		return false
	}
	a := int(bid[0])
	b := int(bid[1])
	fpath := fmt.Sprintf("/%d/%d/%s", a, b, fid)
	basePath := path.Join(conf.Cfg.FileServer.SaveDir, fpath)
	err = os.RemoveAll(basePath)
	if err != nil {
		return false
	}
	fpath = fmt.Sprintf("/%d/%d", a, b)
	basePath = path.Join(conf.Cfg.FileServer.SaveDir, fpath)
	if !deleteEmptyFolder(basePath) {
		return false
	}
	fpath = fmt.Sprintf("/%d", a)
	basePath = path.Join(conf.Cfg.FileServer.SaveDir, fpath)
	if !deleteEmptyFolder(basePath) {
		return false
	}
	return true
}
func isDirEmpty(name string) bool {
	f, err := os.Open(name)
	if err != nil {
		return false
	}
	defer f.Close()
	_, err = f.Readdir(1)
	if err == nil {
		return true
	}
	return false
}
func deleteEmptyFolder(path string) bool {
	empty := isDirEmpty(path)
	if !empty {
		return false
	}
	if err := os.Remove(path); err != nil {
		return false
	}
	return true
}
