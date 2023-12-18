package service

import (
	"database/sql"
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"wios_server/dao"
	"wios_server/entity"
	"wios_server/utils"
)

type ApiService struct {
	*dao.BaseDao
}

func NewApiService(db *sql.DB) *ApiService {
	return &ApiService{BaseDao: dao.NewBaseDao(db)}
}
func (dao *ApiService) ListByPage(pageNumber, pageSize int, fn func(sm *mak.SQLSM)) (int64, []entity.Api, error) {
	var apis []entity.Api
	count, err := dao.ByPage(&apis, pageNumber, pageSize, fn)
	if err != nil {
		return 0, apis, err
	}
	return count, apis, nil
}
func (dao *ApiService) SaveUpdate(api *entity.Api) sql.Result {
	if api.Id == nil {
		*api.Id = utils.UUID()
	}
	return dao.SaveOrUpdate(api)
}
