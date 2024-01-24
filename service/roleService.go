package service

import (
	"database/sql"
	"github.com/CCLooMi/sql-mak/mysql"
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"github.com/sirupsen/logrus"
	"wios_server/conf"
	"wios_server/dao"
	"wios_server/entity"
	"wios_server/handlers/beans"
	"wios_server/utils"
)

type RoleService struct {
	*dao.BaseDao
}

func NewRoleService(db *sql.DB) *RoleService {
	return &RoleService{BaseDao: dao.NewBaseDao(db)}
}
func (dao *RoleService) ListByPage(pageNumber, pageSize int, fn func(sm *mak.SQLSM)) (int64, []entity.Role, error) {
	var roles []entity.Role
	count, err := dao.ByPage(&roles, pageNumber, pageSize, fn)
	if err != nil {
		return 0, roles, err
	}
	return count, roles, nil
}
func (dao *RoleService) SaveUpdate(role *entity.Role) sql.Result {
	if role.Id == nil {
		*role.Id = utils.UUID()
	}
	return dao.SaveOrUpdate(role)
}
func (dao *RoleService) DeleteRole(e *entity.Role) []sql.Result {
	tx, err := conf.Db.Begin()
	if err != nil {
		panic(err.Error())
	}
	dm := mysql.DELETE().FROM(entity.Role{}).
		WHERE("id = ?", e.Id)
	dm2 := mysql.DELETE().FROM(entity.RolePermission{}).
		WHERE("role_id = ?", e.Id)
	dm3 := mysql.DELETE().FROM(entity.RoleMenu{}).
		WHERE("role_id = ?", e.Id)
	dm4 := mysql.DELETE().FROM(entity.RoleUser{}).
		WHERE("role_id = ?", e.Id)
	rs := mysql.TxExecute(tx, dm, dm2, dm3, dm4)
	return rs
}
func (dao *RoleService) FindMenusByRole(e *entity.Role) []beans.MenuWithChecked {
	sm := mysql.SELECT("m.*").
		SELECT_AS("IF(rm.role_id,'on',NULL)", "checked").
		FROM(entity.Menu{}, "m").
		LEFT_JOIN(entity.RoleMenu{}, "rm", "(m.id = rm.menu_id AND rm.role_id = ?)", e.Id)
	var menus []beans.MenuWithChecked
	dao.FindBySM(sm, &menus)
	return menus
}
func (dao *RoleService) FindUsersByRoleId(roleId string, pageNumber int, pageSize int, yes bool) map[string]interface{} {
	var users []entity.User
	//dao.FindBySM(sm, &users)
	count, err := dao.ByPage(&users, pageNumber, pageSize, func(sm *mak.SQLSM) {
		sm.SELECT("DISTINCT u.*").
			FROM(entity.User{}, "u").
			LEFT_JOIN(entity.RoleUser{}, "ru", "ru.user_id = u.id")
		if yes {
			sm.WHERE("ru.role_id = ?", roleId)
		} else {
			sm.WHERE("(ISNULL(ru.role_id) OR ru.role_id <> ?)", roleId)
		}
	})
	if err != nil {
		logrus.Warn(err.Error())
	}
	return map[string]interface{}{
		"total": count,
		"data":  users,
	}
}
func (dao *RoleService) AddMenu(e *entity.RoleMenu) sql.Result {
	return dao.SaveOrUpdate(e)
}
func (dao *RoleService) RemoveMenu(e *entity.RoleMenu) sql.Result {
	return dao.Delete(e)
}
func (dao *RoleService) AddUser(e *entity.RoleUser) sql.Result {
	return dao.SaveOrUpdate(e)
}
func (dao *RoleService) RemoveUser(e *entity.RoleUser) sql.Result {
	return dao.Delete(e)
}
func (dao *RoleService) UpdateMenus(add []entity.RoleMenu, del []interface{}) []sql.Result {
	tx, err := conf.Db.Begin()
	if err != nil {
		panic(err.Error())
	}
	batchArgs := make([][]interface{}, 0)
	for _, v := range add {
		batchArgs = append(batchArgs, []interface{}{v.Id, v.RoleId, v.MenuId, nil, nil})
	}
	a := make([]mak.SQLMak, 0)

	if len(del) > 0 {
		a = append(a, mysql.DELETE().FROM(entity.RoleMenu{}).
			WHERE_IN("id", del...))
	}
	if len(batchArgs) > 0 {
		a = append(a, mysql.INSERT_INTO(entity.RoleMenu{}).
			SetBatchArgs(batchArgs...))
	}
	rs := mysql.TxExecute(tx, a...)
	return rs
}
func (dao *RoleService) FindPermissionsByRole(e *entity.Role) map[string]bool {
	sm := mysql.SELECT("rp.permission_id").
		FROM(entity.RolePermission{}, "rp").
		WHERE("rp.role_id = ?", e.Id)
	var ps []string
	dao.FindBySM(sm, &ps)
	psMap := make(map[string]bool)
	for _, v := range ps {
		psMap[v] = true
	}
	return psMap
}
