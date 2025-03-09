package service

import (
	"database/sql"
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"wios_server/dao"
	"wios_server/entity"
	"wios_server/utils"
)

type PlaylistService struct {
	*dao.BaseDao
	db *sql.DB
}

func NewPlaylistService(db *sql.DB) *PlaylistService {
	return &PlaylistService{BaseDao: dao.NewBaseDao(db), db: db}
}

func (dao *PlaylistService) ListByPage(pageNumber, pageSize int, fn func(sm *mak.SQLSM)) (int64, []entity.Playlist, error) {
	var playlists []entity.Playlist
	count, err := dao.ByPage(&playlists, pageNumber, pageSize, fn)
	if err != nil {
		return 0, playlists, err
	}
	return count, playlists, nil
}

func (dao *PlaylistService) SaveUpdate(playlist *entity.Playlist) sql.Result {
	if playlist.Id == nil {
		id := utils.UUID()
		playlist.Id = &id
	}
	return dao.SaveOrUpdate(playlist)
}

func (dao *PlaylistService) SaveUpdates(playlists []entity.Playlist) []sql.Result {
	list := make([]interface{}, len(playlists))
	for i := 0; i < len(playlists); i++ {
		if playlists[i].Id == nil {
			id := utils.UUID()
			playlists[i].Id = &id
		}
		list[i] = &playlists[i]
	}
	return dao.BatchSaveOrUpdate(list...)
}
