package handlers

import (
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"github.com/gin-gonic/gin"
	"wios_server/conf"
	"wios_server/entity"
	"wios_server/handlers/msg"
	"wios_server/middlewares"
	"wios_server/service"
)

type ConfigController struct {
	configService *service.ConfigService
}

func NewConfigController(app *gin.Engine) *ConfigController {
	ctrl := &ConfigController{configService: service.NewConfigService(conf.Db)}
	group := app.Group("/config")
	hds := []middlewares.Auth{
		{Method: "POST", Group: "/config", Path: "/byPage", Auth: "config.list", Handler: ctrl.byPage},
		{Method: "POST", Group: "/config", Path: "/saveUpdate", Auth: "config.update", Handler: ctrl.saveUpdate},
		{Method: "POST", Group: "/config", Path: "/delete", Auth: "config.delete", Handler: ctrl.delete},
		{Method: "GET", Group: "/config", Path: "/reload", Auth: "config.reload", Handler: ctrl.reloadSysConfig},
	}
	for i, hd := range hds {
		middlewares.RegisterAuth(&hds[i])
		group.Handle(hd.Method, hd.Path, hd.Handler)
	}
	return ctrl
}
func (ctrl *ConfigController) byPage(ctx *gin.Context) {
	middlewares.ByPage(ctx, func(page *middlewares.Page) (int64, any, error) {
		return ctrl.configService.ListByPage(page.PageNumber, page.PageSize, func(sm *mak.SQLSM) {
			sm.SELECT("*").FROM(entity.Config{}, "c")
		})
	})
}
func (ctrl *ConfigController) saveUpdate(ctx *gin.Context) {
	var config entity.Config
	if err := ctx.ShouldBindJSON(&config); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	var rs = ctrl.configService.SaveUpdate(&config)
	_, err := rs.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	msg.Ok(ctx, &config)
}

func (ctrl *ConfigController) delete(ctx *gin.Context) {
	var config entity.Config
	if err := ctx.ShouldBindJSON(&config); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	var rs = ctrl.configService.Delete(&config)
	affected, err := rs.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	if affected > 0 {
		msg.Ok(ctx, &config)
		return
	}
	msg.Error(ctx, "delete failed")
}
func (ctrl *ConfigController) reloadSysConfig(ctx *gin.Context) {
	conf.LoadSysCfg(conf.Db)
	msg.Ok(ctx, "ok")
}
