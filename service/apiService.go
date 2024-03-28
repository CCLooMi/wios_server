package service

import (
	"database/sql"
	"encoding/csv"
	"github.com/CCLooMi/sql-mak/mysql"
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"os"
	"wios_server/conf"
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
	err := os.MkdirAll("static/bak", os.ModePerm)
	if err != nil {
		return err
	}
	file, err := os.Create("static/bak/api.backup.csv")
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()

	page := 0
	pgSize := 1000
	for {
		data := mysql.SELECT("*").
			FROM(entity.Api{}, "a").
			LIMIT(page*(pgSize+1), pgSize).
			Execute(conf.Db).
			GetResultAsCSVData()
		if len(data) > 1 {
			err := writer.WriteAll(data)
			if err != nil {
				return err
			}
			if len(data) < pgSize {
				break
			}
			page++
			continue
		}
		break
	}
	return nil
}
