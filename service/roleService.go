package service

import (
	"database/sql"
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"wios_server/dao"
	"wios_server/entity"
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
