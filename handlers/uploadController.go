package handlers

import (
	"database/sql"
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"github.com/gin-gonic/gin"
	"wios_server/entity"
	"wios_server/handlers/msg"
	"wios_server/middlewares"
	"wios_server/service"
	"wios_server/utils"
)

type UploadController struct {
	uploadService *service.UploadService
	utils         *utils.Utils
}

func NewUploadController(app *gin.Engine, db *sql.DB, utils *utils.Utils) *UploadController {
	ctrl := &UploadController{
		uploadService: service.NewUploadService(db),
		utils:         utils,
	}
	group := app.Group("/upload")
	hds := []middlewares.Auth{
		{Method: "POST", Group: "/upload", Path: "/byPage", Auth: "upload.list", Handler: ctrl.byPage},
		{Method: "POST", Group: "/upload", Path: "/saveUpdate", Auth: "upload.update", Handler: ctrl.saveUpdate},
		{Method: "POST", Group: "/upload", Path: "/delete", Auth: "upload.delete", Handler: ctrl.delete},
	}
	for i, hd := range hds {
		middlewares.RegisterAuth(&hds[i])
		group.Handle(hd.Method, hd.Path, hd.Handler)
	}
	return ctrl
}

func (ctrl *UploadController) byPage(ctx *gin.Context) {
	middlewares.ByPage(ctx, func(page *middlewares.Page) (int64, any, error) {
		return ctrl.uploadService.ListByPage(page.PageNumber, page.PageSize, func(sm *mak.SQLSM) {
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
	_, err := rs.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	msg.Ok(ctx, &upload)
}
func (ctrl *UploadController) delete(ctx *gin.Context) {
	var upload entity.Upload
	if err := ctx.ShouldBindJSON(&upload); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	if !ctrl.utils.DelFileByFid(*upload.Id) {
		msg.Error(ctx, "delete file dir failed")
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
