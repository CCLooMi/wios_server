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
	ut *utils.Utils
}

func NewApiService(db *sql.DB, ut *utils.Utils) *ApiService {
	return &ApiService{BaseDao: dao.NewBaseDao(db), ut: ut}
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

func (dao *ApiService) SaveUpdates(apis []entity.Api) []sql.Result {
	list := make([]interface{}, len(apis))
	for i := 0; i < len(apis); i++ {
		if apis[i].Id == nil {
			*apis[i].Id = utils.UUID()
		}
		list[i] = &apis[i]
	}
	return dao.BatchSaveOrUpdate(list...)
}

func (dao *ApiService) Backup() error {
	a := entity.Api{}
	return dao.ut.BackupTableDataToCSV(
		a.TableName(),
		"static/bak",
		"api.backup.csv")
}
