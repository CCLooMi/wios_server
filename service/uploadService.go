package service

import (
	"database/sql"
	"github.com/CCLooMi/sql-mak/mysql"
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"wios_server/dao"
	"wios_server/entity"
	"wios_server/utils"
)

type UploadService struct {
	*dao.BaseDao
	db *sql.DB
}

func NewUploadService(db *sql.DB) *UploadService {
	return &UploadService{BaseDao: dao.NewBaseDao(db), db: db}
}
func (dao *UploadService) ListByPage(pageNumber, pageSize int, fn func(sm *mak.SQLSM)) (int64, []entity.Upload, error) {
	var uploads []entity.Upload
	count, err := dao.ByPage(&uploads, pageNumber, pageSize, fn)
	if err != nil {
		return 0, uploads, err
	}
	return count, uploads, nil
}
func (dao *UploadService) SaveUpdate(upload *entity.Upload) sql.Result {
	if upload.Id == nil {
		*upload.Id = utils.UUID()
	}
	return dao.SaveOrUpdate(upload)
}

func (dao *UploadService) UpdateUploadSize(fid *string, size *int64) sql.Result {
	return mysql.UPDATE(&entity.Upload{}, "u").
		SET("u.upload_size = ?", size).
		WHERE("u.id = ?", fid).
		AND("u.file_size <> u.upload_size").
		Execute(dao.db).
		Update()
}
