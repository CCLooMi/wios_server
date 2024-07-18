package service

import (
	"database/sql"
	"github.com/CCLooMi/sql-mak/mysql"
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"wios_server/dao"
	"wios_server/entity"
	"wios_server/utils"
)

type MenuService struct {
	*dao.BaseDao
	db *sql.DB
}

func NewMenuService(db *sql.DB) *MenuService {
	return &MenuService{BaseDao: dao.NewBaseDao(db), db: db}
}

func (dao *MenuService) ListByPage(pageNumber, pageSize int, fn func(sm *mak.SQLSM)) (int64, []entity.Menu, error) {
	var menus []entity.Menu
	count, err := dao.ByPage(&menus, pageNumber, pageSize, fn)
	if err != nil {
		return 0, menus, err
	}
	return count, menus, nil
}
func (dao *MenuService) DeleteMenu(menu *entity.Menu) []sql.Result {
	// 开启事务
	tx, err := dao.db.Begin()
	if err != nil {
		panic(err.Error())
	}
	sm := mysql.SELECT("m.id").FROM(entity.Menu{}, "m").
		WHERE("m.id = ?", menu.Id).
		OR("m.pid = ?", menu.Id)
	dm := mysql.DELETE().FROM(entity.RoleMenu{}).
		WHERE_SUBQUERY("menu_id", mak.INValue, sm)
	dm2 := mysql.DELETE().FROM(entity.Menu{}).
		WHERE("id = ?", menu.Id).
		OR("pid = ?", menu.Id)
	rs := mysql.TxExecute(tx, dm, dm2)
	return rs
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
