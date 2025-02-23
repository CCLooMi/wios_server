package service

import (
	"database/sql"
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"wios_server/dao"
	"wios_server/entity"
	"wios_server/utils"
)

type FilesService struct {
	*dao.BaseDao
	db *sql.DB
}

func NewFilesService(db *sql.DB) *FilesService {
	return &FilesService{BaseDao: dao.NewBaseDao(db), db: db}
}

func (dao *FilesService) ListByPage(pageNumber, pageSize int, fn func(sm *mak.SQLSM)) (int64, []entity.Files, error) {
	var files []entity.Files
	count, err := dao.ByPage(&files, pageNumber, pageSize, fn)
	if err != nil {
		return 0, files, err
	}
	return count, files, nil
}
func (dao *FilesService) SaveUpdate(files *entity.Files) sql.Result {
	if files.Id == nil {
		id := utils.UUID()
		files.Id = &id
	}
	return dao.SaveOrUpdate(files)
}
func (dao *FilesService) SaveUpdates(files []entity.Files) []sql.Result {
	list := make([]interface{}, len(files))
	for i := 0; i < len(files); i++ {
		if files[i].Id == nil {
			id := utils.UUID()
			files[i].Id = &id
		}
		list[i] = &files[i]
	}
	return dao.BatchSaveOrUpdate(list...)
}
