package handlers

import (
	"database/sql"
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"github.com/gin-gonic/gin"
	"wios_server/conf"
	"wios_server/entity"
	"wios_server/handlers/msg"
	"wios_server/middlewares"
	"wios_server/service"
)

type PlaylistController struct {
	playlistService *service.PlaylistService
	config          *conf.Config
}

func NewPlaylistController(app *gin.Engine, db *sql.DB, config *conf.Config) *PlaylistController {
	ctrl := &PlaylistController{
		playlistService: service.NewPlaylistService(db),
		config:          config,
	}
	group := app.Group("/playlist")
	hds := []middlewares.Auth{
		{Method: "POST", Group: "/playlist", Path: "/byPage", Auth: "playlist.list", Handler: ctrl.byPage},
		{Method: "POST", Group: "/playlist", Path: "/saveUpdate", Auth: "playlist.update", Handler: ctrl.saveUpdate},
		{Method: "POST", Group: "/playlist", Path: "/saveUpdates", Auth: "playlist.updates", Handler: ctrl.saveUpdates},
		{Method: "POST", Group: "/playlist", Path: "/delete", Auth: "playlist.delete", Handler: ctrl.delete},
	}
	for i, hd := range hds {
		middlewares.RegisterAuth(&hds[i])
		group.Handle(hd.Method, hd.Path, hd.Handler)
	}
	return ctrl
}

func (ctrl *PlaylistController) byPage(ctx *gin.Context) {
	middlewares.ByPage(ctx, func(page *middlewares.Page) (int64, any, error) {
		return ctrl.playlistService.ListByPage(page.PageNumber, page.PageSize, func(sm *mak.SQLSM) {
			sm.SELECT("*").FROM(entity.Playlist{}, "p")
		})
	})
}
func (ctrl *PlaylistController) saveUpdate(ctx *gin.Context) {
	var playlist entity.Playlist
	if err := ctx.ShouldBindJSON(&playlist); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	userInfo, ok := ctx.Get("userInfo")
	if !ok {
		msg.Error(ctx, "userInfo not found")
		return
	}
	userId := userInfo.(*middlewares.UserInfo).User.Id
	if playlist.Id == nil {
		playlist.InsertedBy = userId
	}
	playlist.UpdatedBy = userId
	var rs = ctrl.playlistService.SaveUpdate(&playlist)
	_, err := rs.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	msg.Ok(ctx, &playlist)
}
func (ctrl *PlaylistController) saveUpdates(ctx *gin.Context) {
	var pList []entity.Playlist
	if err := ctx.ShouldBindJSON(&pList); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	userInfo, ok := ctx.Get("userInfo")
	if !ok {
		msg.Error(ctx, "userInfo not found")
		return
	}
	userId := userInfo.(*middlewares.UserInfo).User.Id
	for i := range pList {
		if pList[i].Id == nil {
			pList[i].InsertedBy = userId
		}
		pList[i].UpdatedBy = userId
	}
	var rs = ctrl.playlistService.SaveUpdates(pList)
	for _, r := range rs {
		_, err := r.RowsAffected()
		if err != nil {
			msg.Error(ctx, err.Error())
			return
		}
	}
	msg.Ok(ctx, pList)
}
func (ctrl *PlaylistController) delete(ctx *gin.Context) {
	var playlist entity.Playlist
	if err := ctx.ShouldBindJSON(&playlist); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	var rs = ctrl.playlistService.Delete(&playlist)
	_, err := rs.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	msg.Ok(ctx, &playlist)
}
