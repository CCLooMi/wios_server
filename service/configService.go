package service

import (
	"database/sql"
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"wios_server/dao"
	"wios_server/entity"
	"wios_server/utils"
)

type ConfigService struct {
	*dao.BaseDao
}

func NewConfigService(db *sql.DB) *ConfigService {
	return &ConfigService{BaseDao: dao.NewBaseDao(db)}
}

func (dao *ConfigService) ListByPage(pageNumber, pageSize int, fn func(sm *mak.SQLSM)) (int64, []entity.Config, error) {
	var configs []entity.Config
	count, err := dao.ByPage(&configs, pageNumber, pageSize, fn)
	if err != nil {
		return 0, configs, err
	}
	return count, configs, nil
}

func (dao *ConfigService) SaveUpdate(config *entity.Config) sql.Result {
	if config.Id == nil {
		*config.Id = utils.UUID()
	}
	return dao.SaveOrUpdate(config)
}
