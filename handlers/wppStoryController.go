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

type WppStoryController struct {
	wppStoryService *service.WppStoryService
}

func NewWppStoryController(app *gin.Engine) *WppStoryController {
	ctrl := &WppStoryController{
		wppStoryService: service.NewWppStoryService(conf.Db),
	}
	group := app.Group("/wstory")
	hds := []middlewares.Auth{
		{Method: "POST", Group: "/wstory", Path: "/byPage", Handler: ctrl.byPage},
		{Method: "POST", Group: "/wstory", Path: "/saveUpdate", Auth: "wstory.update", Handler: ctrl.saveUpdate},
		{Method: "POST", Group: "/wstory", Path: "/delete", Auth: "wstory.delete", Handler: ctrl.delete},
	}
	for i, hd := range hds {
		middlewares.RegisterAuth(&hds[i])
		group.Handle(hd.Method, hd.Path, hd.Handler)
	}
	return ctrl
}
func (ctrl *WppStoryController) byPage(c *gin.Context) {
	middlewares.ByPage(c, func(page *middlewares.Page) (int64, any, error) {
		return ctrl.wppStoryService.ListByPage(page.PageNumber, page.PageSize, func(sm *mak.SQLSM) {
			sm.SELECT("*").FROM(entity.WppStory{}, "ws").
				ORDER_BY("ws.updated_at DESC")
		})
	})
}
func (ctrl *WppStoryController) saveUpdate(c *gin.Context) {
	var wppStory entity.WppStory
	if err := c.ShouldBindJSON(&wppStory); err != nil {
		msg.Error(c, err.Error())
		return
	}
	userInfo := c.MustGet(middlewares.UserInfoKey).(*middlewares.UserInfo)
	wppStory.UpdatedBy = userInfo.User.Id
	if wppStory.InsertedBy == nil {
		wppStory.InsertedBy = userInfo.User.Id
	}
	if wppStory.UpdatedBy != nil {
		wppStory.UpdatedAt = nil
	}
	var rs = ctrl.wppStoryService.SaveUpdateWithFilter(&wppStory, func(fieldName *string, columnName *string, v interface{}, im *mak.SQLIM) bool {
		if utils.IsNil(v) {
			return false
		}
		return true
	})
	_, err := rs.RowsAffected()
	if err != nil {
		msg.Error(c, err.Error())
		return
	}
	msg.Ok(c, &wppStory)
}
func (ctrl *WppStoryController) delete(c *gin.Context) {
	var wppStory entity.WppStory
	if err := c.ShouldBindJSON(&wppStory); err != nil {
		msg.Error(c, err.Error())
		return
	}
	var rs = ctrl.wppStoryService.Delete(&wppStory)
	affected, err := rs.RowsAffected()
	if err != nil {
		msg.Error(c, err.Error())
		return
	}
	if affected > 0 {
		msg.Ok(c, &wppStory)
		return
	}
	msg.Error(c, "delete failed")
}
