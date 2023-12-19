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

type UploadController struct {
	uploadService *service.UploadService
}

func NewUploadController(app *gin.Engine) *UploadController {
	ctrl := &UploadController{uploadService: service.NewUploadService(conf.Db)}
	group := app.Group("/upload")
	hds := []middlewares.Auth{
		{Method: "POST", Group: "/upload", Path: "/byPage", Auth: "upload.list", Handler: ctrl.byPage},
		{Method: "POST", Group: "/upload", Path: "/saveUpdate", Auth: "upload.update", Handler: ctrl.saveUpdate},
		{Method: "POST", Group: "/upload", Path: "/delete", Auth: "upload.delete", Handler: ctrl.delete},
	}
	for i, hd := range hds {
		middlewares.AuthMap[hd.Group+hd.Path] = &hds[i]
		group.Handle(hd.Method, hd.Path, hd.Handler)
	}
	return ctrl
}

func (ctrl *UploadController) byPage(ctx *gin.Context) {
	middlewares.ByPage(ctx, func(pageNumber int, pageSize int) (int64, any, error) {
		return ctrl.uploadService.ListByPage(pageNumber, pageSize, func(sm *mak.SQLSM) {
			sm.SELECT("*").FROM(entity.Upload{}, "u")
		})
	})
}

func (ctrl *UploadController) saveUpdate(ctx *gin.Context) {
	var upload entity.Upload
	if err := ctx.ShouldBindJSON(&upload); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	var rs = ctrl.uploadService.SaveUpdate(&upload)
	affected, err := rs.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	if affected > 0 {
		msg.Ok(ctx, &upload)
		return
	}
	msg.Error(ctx, "saveUpdate failed")
}
func (ctrl *UploadController) delete(ctx *gin.Context) {
	var upload entity.Upload
	if err := ctx.ShouldBindJSON(&upload); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	var rs = ctrl.uploadService.Delete(&upload)
	affected, err := rs.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	if affected > 0 {
		msg.Ok(ctx, &upload)
		return
	}
	msg.Error(ctx, "delete failed")
}