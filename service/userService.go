package service

import (
	"database/sql"
	"github.com/CCLooMi/sql-mak/mysql"
	"wios_server/dao"
	"wios_server/entity"
	"wios_server/utils"

	"github.com/CCLooMi/sql-mak/mysql/mak"
)

type UserService struct {
	*dao.BaseDao
}

func NewUserService(db *sql.DB) *UserService {
	return &UserService{BaseDao: dao.NewBaseDao(db)}
}

func (dao *UserService) FindById(id uint) (*entity.User, error) {
	var user entity.User
	dao.ById(id, &user)
	return &user, nil
}

func (dao *UserService) ListByPage(pageNumber, pageSize int, fn func(sm *mak.SQLSM)) (int64, []entity.User, error) {
	var users []entity.User
	count, err := dao.ByPage(&users, pageNumber, pageSize, fn)
	if err != nil {
		return 0, users, err
	}
	return count, users, nil
}
func (dao *UserService) SaveUpdate(user *entity.User) sql.Result {
	if user.Id == nil {
		*user.Id = utils.UUID()
	}
	return dao.SaveOrUpdate(user)
}

func (dao *UserService) FindByUsernameAndPassword(username string, password string) *entity.User {
	var user entity.User
	sm := mysql.SELECT("*").
		FROM(user, "u").
		WHERE("u.username = ?", username).
		AND("u.password = SHA2(CONCAT(u.username,?,u.seed),256)", password).
		LIMIT(1)
	dao.FindBySM(sm, &user)
	if user.Id == nil {
		return nil
	}
	return &user
}

func (dao *UserService) CheckExist(e *entity.User) bool {
	var user entity.User
	sm := mysql.SELECT("*").
		FROM(user, "e").
		WHERE("e.username = ?", e.Username).
		LIMIT(1)
	dao.FindBySM(sm, &user)
	return user.Id != nil
}
