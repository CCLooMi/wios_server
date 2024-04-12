package service

import (
	"database/sql"
	"github.com/CCLooMi/sql-mak/mysql"
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"wios_server/dao"
	"wios_server/entity"
	"wios_server/utils"
)

type WppEventService struct {
	*dao.BaseDao
}

func NewWppEventService(db *sql.DB) *WppEventService {
	return &WppEventService{BaseDao: dao.NewBaseDao(db)}
}
func (dao *WppEventService) ListByPage(pageNumber, pageSize int, fn func(sm *mak.SQLSM)) (int64, []entity.WppEvent, error) {
	var wpps []entity.WppEvent
	count, err := dao.ByPage(&wpps, pageNumber, pageSize, fn)
	if err != nil {
		return 0, wpps, err
	}
	return count, wpps, nil
}
func (dao *WppEventService) SaveUpdate(wpp *entity.WppEvent) sql.Result {
	if wpp.Id == nil {
		id := utils.UUID()
		wpp.Id = &id
	}
	return dao.SaveUpdateWithFilter(wpp, func(fieldName *string, columnName *string, v interface{}, im *mak.SQLIM) bool {
		if utils.IsNil(v) {
			return false
		}
		return true
	})
}
func (dao *WppEventService) SaveUpdates(wpps []entity.WppEvent) []sql.Result {
	list := make([]interface{}, len(wpps))
	for i := 0; i < len(wpps); i++ {
		if wpps[i].Id == nil {
			id := utils.UUID()
			wpps[i].Id = &id
		}
		list[i] = &wpps[i]
	}
	return dao.BatchSaveOrUpdate(list...)
}
func (dao *WppEventService) Inactivate(id *string) sql.Result {
	um := mysql.UPDATE(&entity.WppEvent{}, "e").
		SET("e.status = 'inactive'").
		WHERE("e.id = ?", id)
	return dao.ExecuteUm(um).Update()
}
