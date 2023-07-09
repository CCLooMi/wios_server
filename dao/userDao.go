package dao

import (
	"database/sql"
	"wios_server/entity"
)

type UserDao struct {
	*BaseDao
}

func NewUserDao(db *sql.DB) *UserDao {
	return &UserDao{BaseDao: NewBaseDao(db)}
}

func (dao *UserDao) FindById(id uint) (*entity.User, error) {
	var user entity.User
	dao.ById(id, &user)
	return &user, nil
}

func (dao *UserDao) FindByPage(pageNumber, pageSize int) (int64, []entity.User, error) {
	var users []entity.User
	count, err := dao.ByPage(&users, pageNumber, pageSize)
	if err != nil {
		return 0, nil, err
	}
	return count, users, nil
}
