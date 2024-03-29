package handlers

import (
	"github.com/CCLooMi/sql-mak/mysql"
	"github.com/CCLooMi/sql-mak/mysql/mak"
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

var VMFuncs = make(map[string]interface{})

func init() {
	set("lookupDNSRecord", utils.LookupDNSRecord)
	set("openExcelById", utils.OpenExcelByFid)
	set("setSheetRow", utils.SetExcelSheetRow)
	set("setSheetRows", utils.SetExcelSheetRows)
	set("cellNameToCoordinates", utils.CellNameToCoordinates)
	set("coordinatesToCellName", utils.CoordinatesToCellName)
	set("delFileById", utils.DelFileByFid)
	set("sendEmail", utils.SendEmail)
	set("UUID", utils.UUID)
	set("uuid", utils.UUID)
	set("db", conf.Db)
	set("rdb", conf.Rdb)
	set("cfg", conf.Cfg)
	set("sql", mysqlM)
	set("exp", expM)
}
func set(key string, value interface{}) {
	VMFuncs[key] = value
}
