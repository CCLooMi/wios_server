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
	return dao.SaveUpdateWithFilter(wpp, func(fieldName *string, columnName *string, v interface{}, im *mak.SQLIM) bool {
		if utils.IsNil(v) {
			return false
		}
		return true
	})
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

func (dao *WppService) TopWpps(q string, t int, limit int) []map[string]interface{} {
	sm := mysql.SELECT("*").
		FROM(entity.Wpp{}, "w")
	if q != "" {
		sm.WHERE("w.name LIKE ?", "%"+q+"%")
	}
	switch t {
	case 0:
		sm.ORDER_BY("w.download_count DESC")
	case 1:
		sm.ORDER_BY("w.rating DESC")
	case 2:
		sm.ORDER_BY("w.comment_count DESC")
	case 3:
		sm.ORDER_BY("w.updated_at DESC")
	}
	sm.LIMIT(limit)
	return dao.ExecuteSM(sm).GetResultAsMapList()
}

func (dao *WppService) IsWpp(fid *string) *string {
	sm := mysql.SELECT("rn.wpp_id").
		FROM(&entity.ReleaseNote{}, "rn").
		WHERE("rn.file_id = ?", fid).
		LIMIT(1)
	var s *string
	dao.ExecuteSM(sm).ExtractorResultSet(func(rs *sql.Rows) interface{} {
		for rs.Next() {
			if rs.Scan(s) != nil {
				return nil
			}
			return s
		}
		return nil
	})
	return s
}
func (dao *WppService) PurchaseWpp(wppId *string, userId *string, forcePurchase bool) sql.Result {
	sm := mysql.
		SELECT_EXP_AS(mak.ExpStr("IFNULL(tw.id,REPLACE(UUID(), '-', ''))"), "id").
		SELECT_AS("su.id", "wpp_id").
		SELECT_AS("w.id", "wpp_id").
		SELECT("w.price").
		SELECT_AS("NOW()", "purchase_time").
		SELECT_AS("NOW()", "inserted_at").
		SELECT_AS("NOW()", "updated_at").
		FROM(&entity.Wpp{}, "w").
		LEFT_JOIN(&entity.StoreUser{}, "su", "su.id=?", userId).
		LEFT_JOIN(&entity.PurchasedWpp{}, "tw", "(tw.user_id = su.id AND tw.wpp_id = w.id)").
		WHERE("w.id = ?", wppId).
		AND("su.id IS NOT NULL")
	if !forcePurchase {
		sm.AND("w.price=0")
	}
	return dao.ExecuteIM(
		mysql.INSERT_INTO(&entity.PurchasedWpp{}).
			VALUES_SM(sm).
			ON_DUPLICATE_KEY_UPDATE().
			SET("updated_at = VALUES(updated_at)"),
	).Update()
}

func (dao *WppService) CheckPurchased(wppId *string, userId *string) bool {
	r := dao.PurchaseWpp(wppId, userId, false)
	c, err := r.RowsAffected()
	if err != nil {
		return false
	}
	if c > 0 {
		return true
	}
	sm := mysql.SELECT().
		FROM(&entity.PurchasedWpp{}, "p").
		WHERE("p.user_id = ?", userId).
		AND("p.wpp_id = ?", wppId)
	return dao.ExecuteSM(sm).Count() > 0
}
