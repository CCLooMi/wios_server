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
}

func (dao *BaseDao) Create(entity interface{}) *sql.Result {
	return nil
}

func (dao *BaseDao) Update(entity interface{}) *sql.Result {
	return nil

}

func (dao *BaseDao) Delete(entity interface{}) *sql.Result {
	return nil

}

func (dao *BaseDao) ByPage(out interface{}, pageNumber, pageSize int) (int64, error) {
	outType := utils.GetType(reflect.TypeOf(out))
	fmt.Println(outType)
	//如果outType为切片
	if outType.Kind() == reflect.Slice {
		outType = utils.GetType(outType.Elem())
		fmt.Println(outType)
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
	sm := mysql.SELECT("*").FROM(outEle, "o").LIMIT(pageNumber*pageSize, pageSize)
	sm.Execute(dao.db).ExtractorResultTo(out)

	if pageNumber == 0 {
		return sm.Execute(dao.db).Count(), nil
	}
	return 0, nil
}
