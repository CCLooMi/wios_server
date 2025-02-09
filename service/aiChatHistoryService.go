package service

import (
	"database/sql"
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"wios_server/dao"
	"wios_server/entity"
	"wios_server/utils"
)

type AiChatHistoryService struct {
	*dao.BaseDao
	ut *utils.Utils
}

func NewAiChatHistoryService(db *sql.DB, ut *utils.Utils) *AiChatHistoryService {
	return &AiChatHistoryService{BaseDao: dao.NewBaseDao(db), ut: ut}
}

func (dao *AiChatHistoryService) ListByPage(pageNumber, pageSize int, fn func(sm *mak.SQLSM)) (int64, []entity.AiChatHistory, error) {
	var aiChatHistories []entity.AiChatHistory
	count, err := dao.ByPage(&aiChatHistories, pageNumber, pageSize, fn)
	if err != nil {
		return 0, aiChatHistories, err
	}
	return count, aiChatHistories, nil
}

func (dao *AiChatHistoryService) SaveUpdate(aiChatHistory *entity.AiChatHistory) sql.Result {
	if aiChatHistory.Id == nil {
		*aiChatHistory.Id = utils.UUID()
	}
	return dao.SaveOrUpdate(aiChatHistory)
}

func (dao *AiChatHistoryService) SaveUpdates(aiChatHistories []entity.AiChatHistory) []sql.Result {
	list := make([]interface{}, len(aiChatHistories))
	for i := 0; i < len(aiChatHistories); i++ {
		if aiChatHistories[i].Id == nil {
			*aiChatHistories[i].Id = utils.UUID()
		}
		list[i] = &aiChatHistories[i]
	}
	return dao.BatchSaveOrUpdate(list...)
}
