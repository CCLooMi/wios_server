package handlers

import (
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"github.com/gin-gonic/gin"
	"wios_server/conf"
	"wios_server/entity"
	"wios_server/handlers/msg"
	"wios_server/middlewares"
	"wios_server/service"
	"wios_server/utils"
)

type WppEventController struct {
	wppEventSercie *service.WppEventService
}

func NewWppEventController(app *gin.Engine) *WppEventController {
	ctrl := &WppEventController{
		wppEventSercie: service.NewWppEventService(conf.Db),
	}
	group := app.Group("/wevent")
	hds := []middlewares.Auth{
		{Method: "POST", Group: "/wevent", Path: "/byPage", Handler: ctrl.byPage},
		{Method: "POST", Group: "/wevent", Path: "/saveUpdate", Auth: "wevent.update", Handler: ctrl.saveUpdate},
		{Method: "POST", Group: "/wevent", Path: "/delete", Auth: "wevent.delete", Handler: ctrl.delete},
	}
	for i, hd := range hds {
		middlewares.RegisterAuth(&hds[i])
		group.Handle(hd.Method, hd.Path, hd.Handler)
	}
	return ctrl
}
func (ctrl *WppEventController) byPage(ctx *gin.Context) {
	middlewares.ByPage(ctx, func(page *middlewares.Page) (int64, any, error) {
		return ctrl.wppEventSercie.ListByPage(page.PageNumber, page.PageSize, func(sm *mak.SQLSM) {
			sm.SELECT("*").FROM(entity.WppEvent{}, "we").
				ORDER_BY("we.updated_at DESC")
		})
	})
}
func (ctrl *WppEventController) saveUpdate(ctx *gin.Context) {
	var wppEvent entity.WppEvent
	if err := ctx.ShouldBindJSON(&wppEvent); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	userInfo := ctx.MustGet(middlewares.UserInfoKey).(*middlewares.UserInfo)
	wppEvent.UpdatedBy = userInfo.User.Id
	if wppEvent.InsertedBy == nil {
		wppEvent.InsertedBy = userInfo.User.Id
	}
	if wppEvent.UpdatedAt != nil {
		wppEvent.UpdatedAt = nil
	}
	var rs = ctrl.wppEventSercie.SaveUpdateWithFilter(&wppEvent, func(fieldName *string, columnName *string, v interface{}, im *mak.SQLIM) bool {
		if utils.IsNil(v) {
			return false
		}
		return true
	})
	_, err := rs.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	msg.Ok(ctx, &wppEvent)
}
func (ctrl *WppEventController) delete(ctx *gin.Context) {
	var wppEvent entity.WppEvent
	if err := ctx.ShouldBindJSON(&wppEvent); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	var rs = ctrl.wppEventSercie.Delete(&wppEvent)
	affected, err := rs.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	if affected > 0 {
		msg.Ok(ctx, &wppEvent)
		return
	}
	msg.Error(ctx, "delete failed")
}
