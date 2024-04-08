package service

import (
	"database/sql"
	"github.com/CCLooMi/sql-mak/mysql"
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"wios_server/dao"
	"wios_server/entity"
	"wios_server/utils"
)

type ReleaseNoteService struct {
	*dao.BaseDao
}

func NewReleaseNoteService(db *sql.DB) *ReleaseNoteService {
	return &ReleaseNoteService{BaseDao: dao.NewBaseDao(db)}
}
func (dao *ReleaseNoteService) ListByPage(pageNumber, pageSize int, fn func(sm *mak.SQLSM)) (int64, []entity.ReleaseNote, error) {
	var wrns []entity.ReleaseNote
	count, err := dao.ByPage(&wrns, pageNumber, pageSize, fn)
	if err != nil {
		return 0, wrns, err
	}
	return count, wrns, nil

}
func (dao *ReleaseNoteService) SaveUpdate(wrn *entity.ReleaseNote) sql.Result {
	if wrn.Id == nil {
		id := utils.UUID()
		wrn.Id = &id
	}
	return dao.SaveOrUpdate(wrn)
}

func (dao *ReleaseNoteService) SaveUpdates(wrns []entity.ReleaseNote) []sql.Result {
	list := make([]interface{}, len(wrns))
	for i := 0; i < len(wrns); i++ {
		if wrns[i].Id == nil {
			id := utils.UUID()
			wrns[i].Id = &id
		}
		list[i] = &wrns[i]
	}
	return dao.BatchSaveOrUpdate(list...)
}
func (dao *ReleaseNoteService) GetLatestVersion(wppId *string) *string {
	sm := mysql.SELECT_AS("MAX(r.version)", "latestVersion").
		FROM(entity.ReleaseNote{}, "r").
		WHERE("r.wpp_id = ?", wppId)
	return dao.ExecuteSM(sm).ExtractorResultSet(func(rs *sql.Rows) interface{} {
		var s string
		for rs.Next() {
			if rs.Scan(&s) != nil {
				return nil
			}
			break
		}
		return &s
	}).(*string)
}
