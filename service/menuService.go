package service

import (
	"database/sql"
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"wios_server/dao"
	"wios_server/entity"
)

type MenuService struct {
	*dao.BaseDao
}

func NewMenuService(db *sql.DB) *MenuService {
	return &MenuService{BaseDao: dao.NewBaseDao(db)}
}

func (dao *MenuService) ListByPage(pageNumber, pageSize int, fn func(sm *mak.SQLSM)) (int64, []entity.Menu, error) {
	var menus []entity.Menu
	count, err := dao.ByPage(&menus, pageNumber, pageSize, fn)
	if err != nil {
		return 0, menus, err
	}
	return count, menus, nil
}

func (dao *MenuService) SaveUpdate(menu *entity.Menu) sql.Result {
	return dao.SaveOrUpdate(menu)
}
