package dao

import "github.com/jinzhu/gorm"

type BaseDao struct {
	db *gorm.DB
}

func NewBaseDao(db *gorm.DB) *BaseDao {
	return &BaseDao{db: db}
}

func (dao *BaseDao) Create(entity interface{}) error {
	result := dao.db.Create(entity)
	return result.Error
}

func (dao *BaseDao) Update(entity interface{}) error {
	result := dao.db.Save(entity)
	return result.Error
}

func (dao *BaseDao) Delete(entity interface{}) error {
	result := dao.db.Delete(entity)
	return result.Error
}

func (dao *BaseDao) Find(entity interface{}, conditions ...interface{}) error {
	result := dao.db.First(entity, conditions...)
	return result.Error
}

func (dao *BaseDao) ByPage(entity interface{}, pageNumber, pageSize int, conditions ...interface{}) (int64, error) {
	var count int64
	if pageNumber == 1 {
		result := dao.db.Model(entity).Where(toQuery(conditions...)).Count(&count)
		if result.Error != nil {
			return -1, result.Error
		}
	}

	if pageNumber <= 0 {
		pageNumber = 1
	}

	offset := (pageNumber - 1) * pageSize
	result := dao.db.Model(entity).Where(toQuery(conditions...)).Limit(pageSize).Offset(offset).Find(entity)
	if result.Error != nil {
		return -1, result.Error
	}

	if pageNumber == 1 {
		return count, nil
	} else {
		return -1, nil
	}
}

func toQuery(args ...interface{}) interface{} {
	if len(args) == 0 {
		return ""
	}
	if len(args) == 1 {
		return args[0]
	}
	return args
}
