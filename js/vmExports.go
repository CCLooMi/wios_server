package js

import (
	"bytes"
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"github.com/CCLooMi/sql-mak/mysql"
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/robertkrimen/otto"
	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/fx"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
	"wios_server/conf"
	"wios_server/handlers/msg"
	"wios_server/utils"
)

type mySQLStruct struct {
	SELECT        any
	SELECT_EXP    any
	SELECT_AS     any
	SELECT_SM_AS  any
	SELECT_EXP_AS any
	INSERT_INTO   any
	UPDATE        any
	DELETE        any
	TxExecute     any
}

var mysqlM = mySQLStruct{
	mysql.SELECT,
	mysql.SELECT_EXP,
	mysql.SELECT_AS,
	mysql.SELECT_SM_AS,
	mysql.SELECT_EXP_AS,
	mysql.INSERT_INTO,
	mysql.UPDATE,
	mysql.DELETE,
	mysql.TxExecute,
}

type expStruct struct {
	Now    any
	UUID   any
	Exp    any
	ExpStr any
}

var expM = expStruct{
	mak.Now,
	mak.UUID,
	mak.Exp,
	mak.ExpStr,
}

type templateStruct struct {
	Parse any
	Apply any
}

var templateM = templateStruct{
	Apply: func(str string, data map[string]interface{}) (string, error) {
		id := md5.Sum([]byte(str))
		name := hex.EncodeToString(id[:])
		return utils.ApplyTemplate(&str, name, data)
	},
}
var vmFuncs = make(map[string]interface{})

func doRegExports(ut *utils.Utils, config *conf.Config, db *sql.DB) {
	RegExport("openExcelById", ut.OpenExcelByFid)
	RegExport("delFileById", ut.DelFileByFid)
	RegExport("sendMail", ut.SendMail)
	RegExport("sendMailWithFiles", ut.SendMailWithFiles)
	RegExport("db", db)
	RegExport("cfg", config)
	RegExport("sysCfg", config.SysConf)
	RegExport("lookupDNSRecord", utils.LookupDNSRecord)
	RegExport("setSheetRow", utils.SetExcelSheetRow)
	RegExport("setSheetRows", utils.SetExcelSheetRows)
	RegExport("cellNameToCoordinates", utils.CellNameToCoordinates)
	RegExport("coordinatesToCellName", utils.CoordinatesToCellName)
	RegExport("UUID", utils.UUID)
	RegExport("uuid", utils.UUID)
	RegExport("sql", mysqlM)
	RegExport("exp", expM)
	RegExport("template", templateM)
	RegExport("sleep", func(call otto.FunctionCall) otto.Value {
		duration, _ := call.Argument(0).ToInteger()
		time.Sleep(time.Duration(duration) * time.Millisecond)
		return otto.UndefinedValue()
	})
	RegExport("str2bs", func(str string) []byte {
		return []byte(str)
	})
	RegExport("bs2str", func(bs []byte) string {
		return string(bs)
	})
	RegExport("v2bs", func(v interface{}) []byte {
		bs, e := msgpack.Marshal(v)
		if e != nil {
			log.Println(e)
			return nil
		}
		return bs
	})
	RegExport("bs2v", func(bs []byte) interface{} {
		var v interface{}
		e := msgpack.Unmarshal(bs, &v)
		if e != nil {
			log.Println(e)
		}
		return v
	})
	RegExport("context", map[string]interface{}{
		"WithCancel":       context.WithCancel,
		"WithTimeout":      context.WithTimeout,
		"WithDeadline":     context.WithDeadline,
		"WithValue":        context.WithValue,
		"Background":       context.Background,
		"AfterFunc":        context.AfterFunc,
		"WithoutCancel":    context.WithoutCancel,
		"WithCancelCause":  context.WithCancelCause,
		"WithTimeoutCause": context.WithTimeoutCause,
		"Cause":            context.Cause,
		"TODO":             context.TODO,
	})
	RegExport("$", func(str string) *goquery.Document {
		result, error := goquery.NewDocumentFromReader(strings.NewReader(str))
		if error != nil {
			log.Println(error)
			return nil
		}
		return result
	})
	RegExport("fetch", func(url string, opts ...interface{}) map[string]interface{} {
		result, err := fetch(url, opts...)
		if err != nil {
			log.Println(err)
			return nil
		}
		return result
	})
}
func RegExport(key string, value interface{}) {
	vmFuncs[key] = value
}
func ApplyExportsTo(vm *otto.Otto) {
	for key, v := range vmFuncs {
		vm.Set(key, v)
	}
}

type MsgUtil struct {
	Ok  func(msg any)
	Err func(msg any)
	Oks func(msgs ...any)
}

func NewMsgUtil(c *gin.Context) *MsgUtil {
	return &MsgUtil{
		Ok: func(m any) {
			msg.Ok(c, m)
		},
		Err: func(m any) {
			msg.Error(c, m)
		},
		Oks: func(m ...any) {
			msg.Oks(c, m...)
		},
	}
}
func getStrDecode(name string) *encoding.Decoder {
	switch strings.ToUpper(name) {
	case "UTF8", "UTF-8":
		return unicode.UTF8.NewDecoder()
	case "GBK", "GB2312":
		return simplifiedchinese.GBK.NewDecoder()
	case "GB18030":
		return simplifiedchinese.GB18030.NewDecoder()
	case "BIG5":
		return traditionalchinese.Big5.NewDecoder()
	case "ISO-8859-1":
		return charmap.ISO8859_1.NewDecoder()
	case "ISO-8859-2":
		return charmap.ISO8859_2.NewDecoder()
	case "ISO-8859-3":
		return charmap.ISO8859_3.NewDecoder()
	case "ISO-8859-4":
		return charmap.ISO8859_4.NewDecoder()
	case "ISO-8859-5":
		return charmap.ISO8859_5.NewDecoder()
	case "ISO-8859-6":
		return charmap.ISO8859_6.NewDecoder()
	case "ISO-8859-7":
		return charmap.ISO8859_7.NewDecoder()
	case "ISO-8859-8":
		return charmap.ISO8859_8.NewDecoder()
	case "ISO-8859-9":
		return charmap.ISO8859_9.NewDecoder()
	case "ISO-8859-10":
		return charmap.ISO8859_10.NewDecoder()
	case "WINDOWS-1250":
		return charmap.Windows1250.NewDecoder()
	case "WINDOWS-1251":
		return charmap.Windows1251.NewDecoder()
	case "WINDOWS-1252":
		return charmap.Windows1252.NewDecoder()
	case "WINDOWS-1253":
		return charmap.Windows1253.NewDecoder()
	case "WINDOWS-1254":
		return charmap.Windows1254.NewDecoder()
	case "WINDOWS-1255":
		return charmap.Windows1255.NewDecoder()
	case "WINDOWS-1256":
		return charmap.Windows1256.NewDecoder()
	case "WINDOWS-1257":
		return charmap.Windows1257.NewDecoder()
	case "WINDOWS-1258":
		return charmap.Windows1258.NewDecoder()
	case "KOI8-R":
		return korean.EUCKR.NewDecoder()
	case "EUC-JP":
		return japanese.EUCJP.NewDecoder()
	case "ISO-2022-JP":
		return japanese.ISO2022JP.NewDecoder()
	case "UTF-16", "UTF-16BE":
		return unicode.UTF16(unicode.BigEndian, unicode.UseBOM).NewDecoder()
	case "UTF-16LE":
		return unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder()
	}
	return unicode.UTF8.NewDecoder()
}

const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36"

func fetch(url string, o ...interface{}) (map[string]interface{}, error) {
	var opts map[string]interface{}
	if len(o) > 0 {
		opts = o[0].(map[string]interface{})
	} else {
		opts = map[string]interface{}{}
	}
	method, ok := opts["method"].(string)
	if !ok {
		method = "GET"
	}
	var body string
	if bodyInterface, ok := opts["body"]; ok {
		body = bodyInterface.(string)
	}
	req, err := http.NewRequest(method, url, bytes.NewBufferString(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	headers, ok := opts["headers"].(map[string]interface{})
	if ok {
		for k, v := range headers {
			req.Header.Set(k, v.(string))
		}
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	charset, ok := opts["charset"].(string)
	if !ok {
		charset = "UTF8"
	}
	decode := getStrDecode(charset)
	rspBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	dc, err := decode.Bytes(rspBody)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"response":   string(dc),
		"request":    resp.Request,
		"status":     resp.Status,
		"statusCode": resp.StatusCode,
		"header":     resp.Header,
		"cookies": func() []*http.Cookie {
			return resp.Cookies()
		},
	}, nil
}

var Module = fx.Options(
	fx.Invoke(doRegExports),
)
