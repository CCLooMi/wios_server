package service

import (
	"database/sql"
	"github.com/CCLooMi/sql-mak/mysql"
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"wios_server/dao"
	"wios_server/entity"
	"wios_server/utils"
)

type WppStoryService struct {
	*dao.BaseDao
}

func NewWppStoryService(db *sql.DB) *WppStoryService {
	return &WppStoryService{BaseDao: dao.NewBaseDao(db)}
}
func (dao *WppStoryService) ListByPage(pageNumber, pageSize int, fn func(sm *mak.SQLSM)) (int64, []entity.WppStory, error) {
	var wpps []entity.WppStory
	count, err := dao.ByPage(&wpps, pageNumber, pageSize, fn)
	if err != nil {
		return 0, wpps, err
	}
	return count, wpps, nil
}
func (dao *WppStoryService) SaveUpdate(wpp *entity.WppStory) sql.Result {
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
func (dao *WppStoryService) SaveUpdates(wpps []entity.WppStory) []sql.Result {
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
func (dao *WppStoryService) Inactivate(id *string) sql.Result {
	um := mysql.UPDATE(&entity.WppStory{}, "s").
		SET("s.status = 'inactive'").
		WHERE("s.id = ?", id)
	return dao.ExecuteUm(um).Update()
}
