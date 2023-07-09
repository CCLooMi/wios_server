package dao

import (
	"database/sql"

	"github.com/CCLooMi/sql-mak/mysql/mak"
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
	return 0, nil
}
