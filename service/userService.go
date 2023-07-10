package service

import (
	"database/sql"
	"wios_server/dao"
	"wios_server/entity"

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

func (dao *UserService) FindByPage(pageNumber, pageSize int) (int64, []entity.User, error) {
	var users []entity.User
	count, err := dao.ByPage(&users, pageNumber, pageSize, func(sm *mak.SQLSM) {
		sm.SELECT("*").FROM(entity.User{}, "u")
	})
	if err != nil {
		return 0, users, err
	}
	return count, users, nil
}
