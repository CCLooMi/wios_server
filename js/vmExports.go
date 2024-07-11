package js

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"github.com/CCLooMi/sql-mak/mysql"
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"github.com/robertkrimen/otto"
	"time"
	"wios_server/conf"
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

func init() {
	RegExport("lookupDNSRecord", utils.LookupDNSRecord)
	RegExport("openExcelById", utils.OpenExcelByFid)
	RegExport("setSheetRow", utils.SetExcelSheetRow)
	RegExport("setSheetRows", utils.SetExcelSheetRows)
	RegExport("cellNameToCoordinates", utils.CellNameToCoordinates)
	RegExport("coordinatesToCellName", utils.CoordinatesToCellName)
	RegExport("delFileById", utils.DelFileByFid)
	RegExport("sendMail", utils.SendMail)
	RegExport("sendMailWithFiles", utils.SendMailWithFiles)
	RegExport("UUID", utils.UUID)
	RegExport("uuid", utils.UUID)
	RegExport("db", conf.Db)
	RegExport("rdb", conf.Rdb)
	RegExport("cfg", conf.Cfg)
	RegExport("sysCfg", conf.SysCfg)
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
}
func RegExport(key string, value interface{}) {
	vmFuncs[key] = value
}
func ApplyExportsTo(vm *otto.Otto) {
	for key, v := range vmFuncs {
		vm.Set(key, v)
	}
}
