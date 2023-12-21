package service

import (
	"database/sql"
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"wios_server/dao"
	"wios_server/entity"
	"wios_server/utils"
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
	if menu.Id == nil {
		*menu.Id = utils.UUID()
	}
	return dao.SaveOrUpdate(menu)
}

func (dao *MenuService) BatchSaveUpdate(menus ...interface{}) []sql.Result {
	return dao.BatchSaveOrUpdate(menus...)
}
