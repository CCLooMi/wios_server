package dao

import (
	"database/sql"
	"fmt"
	"reflect"

	"github.com/CCLooMi/sql-mak/mysql"
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"github.com/CCLooMi/sql-mak/utils"
)

type BaseDao struct {
	db *sql.DB
}
type ByPage func(sm *mak.SQLSM)

func NewBaseDao(db *sql.DB) *BaseDao {
	return &BaseDao{db: db}
}

func (dao *BaseDao) ById(id interface{}, out interface{}) {
	sm := mysql.SELECT("*").FROM(out, "e").WHERE("e.id = ?", id).LIMIT(1)
	sm.Execute(dao.db).ExtractorResultTo(out)
}

func (dao *BaseDao) SaveOrUpdate(entity interface{}) sql.Result {
	ei := utils.GetEntityInfo(entity)
	im := mysql.INSERT_INTO(entity).ON_DUPLICATE_KEY_UPDATE()
	for _, col := range ei.Columns {
		if col != ei.PrimaryKey {
			im.SET("e."+col+"=?", utils.GetFieldValue(entity, ei.CFMap[col]))
		}
	}
	return im.Execute(dao.db).Update()
}

func (dao *BaseDao) Update(entity interface{}) sql.Result {
	ei := utils.GetEntityInfo(entity)
	um := mysql.UPDATE(entity, "e")
	for _, col := range ei.Columns {
		if col != ei.PrimaryKey {
			um.SET("e."+col+"=?", utils.GetFieldValue(entity, ei.CFMap[col]))
		}
	}
	um.WHERE("e."+ei.PrimaryKey+" = ?", utils.GetFieldValue(entity, ei.CFMap[ei.PrimaryKey]))
	return um.Execute(dao.db).Update()
}

func (dao *BaseDao) Delete(entity interface{}) sql.Result {
	ei := utils.GetEntityInfo(entity)
	dm := mysql.DELETE().FROM(entity).WHERE(ei.PrimaryKey+" = ?", utils.GetFieldValue(entity, ei.CFMap[ei.PrimaryKey]))
	return dm.Execute(dao.db).Update()
}

func (dao *BaseDao) ByPage(out interface{}, pageNumber, pageSize int, byPage ByPage) (int64, error) {
	outType := utils.GetType(reflect.TypeOf(out))
	//如果outType为切片
	if outType.Kind() == reflect.Slice {
		outType = utils.GetType(outType.Elem())
	}
	outEle := reflect.New(outType).Elem().Interface()
	fmt.Println(reflect.TypeOf(outEle))
	if pageNumber <= 0 {
		pageNumber = 0
	} else {
		pageNumber = pageNumber - 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	sm := mak.NewSQLSM()
	byPage(sm)
	sm.LIMIT(pageNumber*pageSize, pageSize)
	sm.Execute(dao.db).ExtractorResultTo(out)
	if pageNumber == 0 {
		return sm.Execute(dao.db).Count(), nil
	}
	return 0, nil
}
