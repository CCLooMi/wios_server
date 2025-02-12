package service

import (
	"database/sql"
	"github.com/CCLooMi/sql-mak/mysql"
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"wios_server/dao"
	"wios_server/entity"
	"wios_server/utils"
)

type AiAssistantService struct {
	*dao.BaseDao
	ut *utils.Utils
}

func NewAiAssistantService(db *sql.DB, ut *utils.Utils) *AiAssistantService {
	return &AiAssistantService{BaseDao: dao.NewBaseDao(db), ut: ut}
}

func (dao *AiAssistantService) ListByPage(pageNumber, pageSize int, fn func(sm *mak.SQLSM)) (int64, []entity.AiAssistant, error) {
	var aiAssistants []entity.AiAssistant
	count, err := dao.ByPage(&aiAssistants, pageNumber, pageSize, fn)
	if err != nil {
		return 0, aiAssistants, err
	}
	return count, aiAssistants, nil
}

func (dao *AiAssistantService) SaveUpdate(aiAssistant *entity.AiAssistant) sql.Result {
	if aiAssistant.Id == nil {
		*aiAssistant.Id = utils.UUID()
	}
	return dao.SaveOrUpdate(aiAssistant)
}

func (dao *AiAssistantService) SaveUpdates(aiAssistants []entity.AiAssistant) []sql.Result {
	list := make([]interface{}, len(aiAssistants))
	for i := 0; i < len(aiAssistants); i++ {
		if aiAssistants[i].Id == nil {
			*aiAssistants[i].Id = utils.UUID()
		}
		list[i] = &aiAssistants[i]
	}
	return dao.BatchSaveOrUpdate(list...)
}
func (dao *AiAssistantService) SetStatus(id interface{}, status string) sql.Result {
	um := mysql.UPDATE(entity.AiAssistant{}, "a").
		SET("a.status = ?", status).
		WHERE("id = ?", id)
	return dao.ExecuteUm(um).Update()
}
