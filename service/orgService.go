package service

import (
	"database/sql"
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"wios_server/dao"
	"wios_server/entity"
	"wios_server/utils"
)

type OrgService struct {
	*dao.BaseDao
}

func NewOrgService(db *sql.DB) *OrgService {
	return &OrgService{BaseDao: dao.NewBaseDao(db)}
}
func (dao *OrgService) ListByPage(pageNumber, pageSize int, fn func(sm *mak.SQLSM)) (int64, []entity.Org, error) {
	var orgs []entity.Org
	count, err := dao.ByPage(&orgs, pageNumber, pageSize, fn)
	if err != nil {
		return 0, orgs, err
	}
	return count, orgs, nil
}
func (dao *OrgService) SaveUpdate(org *entity.Org) sql.Result {
	if org.Id == nil {
		*org.Id = utils.UUID()
	}
	return dao.SaveOrUpdate(org)
}
