package dao

import (
	"wios_server/entity"

	"github.com/jinzhu/gorm"
)

type UserDao struct {
	*BaseDao
}

func NewUserDao(db *gorm.DB) *UserDao {
	return &UserDao{BaseDao: NewBaseDao(db)}
}

func (dao *UserDao) FindById(id uint) (*entity.User, error) {
	var user entity.User
	err := dao.Find(&user, "id = ?", id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (dao *UserDao) FindByPage(pageNumber, pageSize int, conditions ...interface{}) (int64, []entity.User, error) {
	var users []entity.User
	count, err := dao.ByPage(&users, pageNumber, pageSize, conditions...)
	if err != nil {
		return 0, nil, err
	}
	return count, users, nil
}
