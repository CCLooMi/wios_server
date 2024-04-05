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

func (dao *BaseDao) FindBySM(sm *mak.SQLSM, out interface{}) {
	sm.Execute(dao.db).ExtractorResultTo(out)
}
func (dao *BaseDao) ExecuteSM(sm *mak.SQLSM) *mak.MySQLSMExecutor {
	return sm.Execute(dao.db)
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
			v := utils.GetFieldValue(entity, ei.CFMap[col])
			if col == "inserted_at" {
				im.SET("inserted_at=COALESCE(?,inserted_at,NOW())", v)
				continue
			}
			if col == "updated_at" {
				im.SET("updated_at=IFNULL(?, NOW())", v)
				continue
			}
			im.SET("`"+col+"`=?", v)
		}
	}
	return im.Execute(dao.db).Update()
}
func (dao *BaseDao) SaveUpdateWithFilter(entity interface{},
	ft func(fieldName *string, columnName *string, v interface{}, im *mak.SQLIM) bool) sql.Result {
	ei := utils.GetEntityInfo(entity)
	im := mysql.INSERT_INTO(entity).ON_DUPLICATE_KEY_UPDATE()
	for _, col := range ei.Columns {
		if col != ei.PrimaryKey {
			fieldName := ei.CFMap[col]
			v := utils.GetFieldValue(entity, fieldName)
			if col == "inserted_at" {
				im.SET("inserted_at=COALESCE(?,inserted_at,NOW())", v)
				continue
			}
			if col == "updated_at" {
				im.SET("updated_at=IFNULL(?, NOW())", v)
				continue
			}
			if ft(&fieldName, &col, v, im) {
				im.SET("`"+col+"`=?", v)
			}
		}
	}
	return im.Execute(dao.db).Update()
}
func (dao *BaseDao) BatchSaveOrUpdate(entities ...interface{}) []sql.Result {
	if len(entities) == 0 {
		return nil
	}
	entity := entities[0]
	ei := utils.GetEntityInfo(entity)
	im := mysql.INSERT_INTO(entity).ON_DUPLICATE_KEY_UPDATE()
	for _, col := range ei.Columns {
		if col != ei.PrimaryKey {
			if col == "inserted_at" {
				im.SET("inserted_at=COALESCE(?,inserted_at,NOW())")
				continue
			}
			if col == "updated_at" {
				im.SET("updated_at=IFNULL(?, NOW())")
				continue
			}
			im.SET("`" + col + "`=?")
		}
	}

	batchArgs := make([][]interface{}, 0)
	for _, entity := range entities {
		args := make([]interface{}, 0)
		args = append(args, utils.GetFieldValue(entity, ei.CFMap[ei.PrimaryKey]))
		for _, col := range ei.Columns {
			if col != ei.PrimaryKey {
				args = append(args, utils.GetFieldValue(entity, ei.CFMap[col]))
			}
		}
		args = append(args, args[1:]...)
		batchArgs = append(batchArgs, args)
	}
	im.SetBatchArgs(batchArgs...)
	return im.Execute(dao.db).BatchUpdate()
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
func (dao *BaseDao) UpdateWithFilter(entity interface{},
	ft func(fieldName *string, columnName *string, v interface{}, um *mak.SQLUM) bool) sql.Result {
	ei := utils.GetEntityInfo(entity)
	um := mysql.UPDATE(entity, "e")
	for _, col := range ei.Columns {
		if col != ei.PrimaryKey {
			fieldName := ei.CFMap[col]
			v := utils.GetFieldValue(entity, fieldName)
			if ft(&fieldName, &col, v, um) {
				um.SET("e."+col+"=?", v)
			}
		}
	}
	um.WHERE("e."+ei.PrimaryKey+" = ?", utils.GetFieldValue(entity, ei.CFMap[ei.PrimaryKey]))
	return um.Execute(dao.db).Update()
}

func (dao *BaseDao) Delete(entity interface{}) sql.Result {
	ei := utils.GetEntityInfo(entity)
	dm := mysql.DELETE().FROM(entity).
		WHERE(ei.PrimaryKey+" = ?", utils.GetFieldValue(entity, ei.CFMap[ei.PrimaryKey]))
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
	if pageNumber < 0 {
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
	return -1, nil
}
