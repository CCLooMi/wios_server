package service

import (
	"database/sql"
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"wios_server/dao"
	"wios_server/entity"
	"wios_server/utils"
)

type PermissionService struct {
	*dao.BaseDao
}

func NewPermissionService(db *sql.DB) *PermissionService {
	return &PermissionService{BaseDao: dao.NewBaseDao(db)}
}
func (dao *PermissionService) ListByPage(pageNumber, pageSize int, fn func(sm *mak.SQLSM)) (int64, []entity.Permission, error) {
	var permissions []entity.Permission
	count, err := dao.ByPage(&permissions, pageNumber, pageSize, fn)
	if err != nil {
		return 0, permissions, err
	}
	return count, permissions, nil
}
func (dao *PermissionService) SaveUpdate(permission *entity.Permission) sql.Result {
	if permission.Id == nil {
		*permission.Id = utils.UUID()
	}
	return dao.SaveOrUpdate(permission)
}
