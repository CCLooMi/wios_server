package service

import (
	"database/sql"
	"github.com/CCLooMi/sql-mak/mysql"
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"wios_server/dao"
	"wios_server/entity"
	"wios_server/utils"
)

type WppService struct {
	*dao.BaseDao
}

func NewWppService(db *sql.DB) *WppService {
	return &WppService{BaseDao: dao.NewBaseDao(db)}
}
func (dao *WppService) ListByPage(pageNumber, pageSize int, fn func(sm *mak.SQLSM)) (int64, []entity.Wpp, error) {
	var wpps []entity.Wpp
	count, err := dao.ByPage(&wpps, pageNumber, pageSize, fn)
	if err != nil {
		return 0, wpps, err
	}
	return count, wpps, nil
}
func (dao *WppService) FindById(id *string) *entity.Wpp {
	var wpp entity.Wpp
	sm := mysql.SELECT("*").
		FROM(wpp, "w").
		WHERE("w.id = ?", id).
		LIMIT(1)
	dao.ExecuteSM(sm).ExtractorResultTo(&wpp)
	return &wpp
}
func (dao *WppService) SaveUpdate(wpp *entity.Wpp) sql.Result {
	if wpp.Id == nil {
		id := utils.UUID()
		wpp.Id = &id
	}
	return dao.SaveOrUpdate(wpp)
}
func (dao *WppService) SaveUpdates(wpps []entity.Wpp) []sql.Result {
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
func (dao *WppService) IsLatestVersion(wppId *string, version *string) (bool, *string) {
	sm := mysql.SELECT_EXP_AS(mak.ExpStr("?>w.latest_version", version), "isLatest").
		SELECT("w.latest_version").
		FROM(entity.Wpp{}, "w").
		WHERE("w.id = ?", wppId)
	var b bool
	var s string
	dao.ExecuteSM(sm).ExtractorResultSet(func(rs *sql.Rows) interface{} {
		for rs.Next() {
			if rs.Scan(&b, &s) != nil {
				b = true
				return nil
			}
			return nil
		}
		b = true
		return nil
	})
	return b, &s
}
