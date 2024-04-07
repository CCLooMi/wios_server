package service

import (
	"database/sql"
	"github.com/CCLooMi/sql-mak/mysql"
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"wios_server/dao"
	"wios_server/entity"
	"wios_server/utils"
)

type StoreUserService struct {
	*dao.BaseDao
}

func NewStoreUserService(db *sql.DB) *StoreUserService {
	return &StoreUserService{BaseDao: dao.NewBaseDao(db)}
}
func (dao *StoreUserService) FindById(id *string) (*entity.StoreUser, error) {
	var user entity.StoreUser
	dao.ById(id, &user)
	return &user, nil
}

func (dao *StoreUserService) ListByPage(pageNumber, pageSize int, fn func(sm *mak.SQLSM)) (int64, []entity.StoreUser, error) {
	var users []entity.StoreUser
	count, err := dao.ByPage(&users, pageNumber, pageSize, fn)
	if err != nil {
		return 0, users, err
	}
	return count, users, nil
}

func (dao *StoreUserService) SaveUpdate(user *entity.StoreUser) sql.Result {
	if user.Id == nil {
		*user.Id = utils.UUID()
	}
	return dao.SaveOrUpdate(user)
}

func (dao *StoreUserService) FindByUsernameAndPassword(username string, password string) *entity.StoreUser {
	var user entity.StoreUser
	sm := mysql.SELECT("*").
		FROM(user, "u").
		WHERE("(u.username = ? or u.email = ? or u.phone = ?)", username, username, username).
		AND("u.password = SHA2(CONCAT(?,u.seed),256)", password).
		LIMIT(1)
	dao.FindBySM(sm, &user)
	return &user
}

func (dao *StoreUserService) CheckExist(e *entity.StoreUser) bool {
	sm := mysql.SELECT("COUNT(1)").
		FROM(entity.StoreUser{}, "e").
		WHERE("(e.username = ? OR e.email = ? OR e.phone = ?)", e.Username, e.Email, e.Phone).
		LIMIT(1)
	return dao.ExecuteSM(sm).Count() > 0
}
