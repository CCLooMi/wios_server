package utils

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/CCLooMi/sql-mak/mysql"
	"github.com/xuri/excelize/v2"
	"html/template"
	"math/big"
	"net"
	"net/smtp"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
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
func SaveKVToRedis(key string, value string, expire time.Duration) error {
	return conf.Rdb.
		Set(conf.Ctx, key, value, expire).
		Err()
}
func GetValueFromRedis(key string) (string, error) {
	data, err := conf.Rdb.Get(conf.Ctx, key).Result()
	if err != nil {
		return "", err
	}
	return data, nil
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
func SHA_256(password string, seed []byte) string {
	b := sha256.Sum256(append([]byte(password), seed...))
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
func SetExcelSheetRows(f *excelize.File, sheet string, cell string, data ...interface{}) error {
	if err := f.SetSheetRow(sheet, cell, &data); err != nil {
		return err
	}
	return nil
}
func SetExcelSheetRow(f *excelize.File, sheet string, cell string, data []interface{}) error {
	if err := f.SetSheetRow(sheet, cell, &data); err != nil {
		return err
	}
	return nil
}
func CellNameToCoordinates(cell string) (int, int, error) {
	return excelize.CellNameToCoordinates(cell)
}
func CoordinatesToCellName(col int, row int, abs ...bool) (string, error) {
	return excelize.CoordinatesToCellName(col, row, abs...)
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

func RemoveDomainPort(domain string) string {
	parts := strings.Split(domain, ":")
	return parts[0]
}
func BackupTableDataToCSV(tableName string, dir string, fileName string) error {
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}
	file, err := os.Create(path.Join(dir, fileName))
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()

	page := 0
	pgSize := 1000
	for {
		data := mysql.SELECT("*").
			FROM(tableName, "a").
			LIMIT(page*(pgSize+1), pgSize).
			Execute(conf.Db).
			GetResultAsCSVData()
		if len(data) > 1 {
			err := writer.WriteAll(data)
			if err != nil {
				return err
			}
			if len(data) < pgSize {
				break
			}
			page++
			continue
		}
		break
	}
	return nil
}
func SendEmail(subject string, body *string, to ...string) error {
	emailCfg := conf.SysCfg["sys.email"].(map[string]interface{})
	port, ok := emailCfg["smtpPort"].(float64)
	if !ok {
		port = 25
	}
	fromEmail := emailCfg["email"].(string)
	password := emailCfg["password"].(string)
	smptHost := emailCfg["smtp"].(string)
	auth := smtp.PlainAuth("", fromEmail, password, smptHost)
	msg := []byte(
		"From: " + fromEmail + "\n" +
			"To: " + strings.Join(to, ",") + "\n" +
			"Subject: " + subject + "\n" +
			"MIME-version: 1.0;\n" +
			"Content-Type: text/html; charset=\"UTF-8\";\n\n" +
			*body)
	return smtp.SendMail(smptHost+":"+strconv.Itoa(int(port)), auth, fromEmail, to, msg)
}
func ApplyTemplate(text *string, name string, data any) (string, error) {
	t := template.New(name)
	_, err := t.Parse(*text)
	if err != nil {
		return "", err
	}
	var buf strings.Builder
	err = t.Execute(&buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

const digits = "0123456789"

var digits_lenth = big.NewInt(int64(len(digits)))

func GenRandomNum(length int) string {
	result := make([]byte, length)
	for i := range result {
		n, err := rand.Int(rand.Reader, digits_lenth)
		if err != nil {
			panic(err)
		}
		result[i] = digits[n.Int64()]
	}
	return string(result)
}
func IsNil(a interface{}) bool {
	if a == nil {
		return true
	}
	value := reflect.ValueOf(a)
	if value.Kind() == reflect.Ptr {
		return value.IsNil()
	}
	return false
}
func IsBlank(v interface{}) bool {
	if IsNil(v) {
		return true
	}
	if str, ok := v.(string); ok && str != "" {
		return false
	}
	return true
}
