package handlers

import (
	"database/sql"
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"github.com/gin-gonic/gin"
	"strings"
	"wios_server/conf"
	"wios_server/entity"
	"wios_server/handlers/msg"
	"wios_server/middlewares"
	"wios_server/service"
	"wios_server/task"
)

type FilesController struct {
	filesService *service.FilesService
	config       *conf.Config
}

func NewFilesController(app *gin.Engine, db *sql.DB, config *conf.Config) *FilesController {
	ctrl := &FilesController{filesService: service.NewFilesService(db), config: config}
	group := app.Group("/files")
	hds := []middlewares.Auth{
		{Method: "POST", Group: "/files", Path: "/byPage", Auth: "files.list", Handler: ctrl.byPage},
		{Method: "POST", Group: "/files", Path: "/saveUpdate", Auth: "files.update", Handler: ctrl.saveUpdate},
		{Method: "POST", Group: "/files", Path: "/saveUpdates", Auth: "files.updates", Handler: ctrl.saveUpdates},
		{Method: "POST", Group: "/files", Path: "/delete", Auth: "files.delete", Handler: ctrl.delete},
		{Method: "POST", Group: "/files", Path: "/genSubtitle", Auth: "files.genSubtitle", Handler: ctrl.genSubtitle},
	}
	for i, hd := range hds {
		middlewares.RegisterAuth(&hds[i])
		group.Handle(hd.Method, hd.Path, hd.Handler)
	}
	return ctrl
}
func (ctrl *FilesController) byPage(ctx *gin.Context) {
	middlewares.ByPage(ctx, func(page *middlewares.Page) (int64, any, error) {
		return ctrl.filesService.ListByPage(page.PageNumber, page.PageSize, func(sm *mak.SQLSM) {
			sm.SELECT("*").
				FROM(entity.Files{}, "f").
				ORDER_BY("f.inserted_at DESC", "f.id")
		})
	})
}
func (ctrl *FilesController) saveUpdate(ctx *gin.Context) {
	var files entity.Files
	if err := ctx.ShouldBindJSON(&files); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	var rs = ctrl.filesService.SaveUpdate(&files)
	_, err := rs.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	msg.Ok(ctx, &files)
}
func (ctrl *FilesController) saveUpdates(ctx *gin.Context) {
	var files []entity.Files
	if err := ctx.ShouldBindJSON(&files); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	userInfo, ok := ctx.Get("userInfo")
	if !ok {
		msg.Error(ctx, "userInfo not found")
		return
	}
	userId := userInfo.(*middlewares.UserInfo).User.Id
	for i := 0; i < len(files); i++ {
		files[i].UserId = userId
	}
	var rs = ctrl.filesService.SaveUpdates(files)
	for _, r := range rs {
		_, err := r.RowsAffected()
		if err != nil {
			msg.Error(ctx, err.Error())
			return
		}
	}
	msg.Ok(ctx, &files)
}
func (ctrl *FilesController) delete(ctx *gin.Context) {
	var files entity.Files
	if err := ctx.ShouldBindJSON(&files); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	var rs = ctrl.filesService.Delete(&files)
	_, err := rs.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	msg.Ok(ctx, &files)
}
func (ctrl *FilesController) genSubtitle(ctx *gin.Context) {
	var files entity.Files
	if err := ctx.ShouldBindJSON(&files); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	if files.FileId == nil || *files.FileId == "" {
		msg.Error(ctx, "fileId not found")
		return
	}
	if files.FileType == nil || *files.FileType == "" {
		msg.Error(ctx, "fileType not found")
		return
	}
	if strings.HasPrefix(*files.FileType, "video/") {
		err := task.ProcessSubtitle(ctrl.config.FileServer.SaveDir, *files.FileId)
		if err != nil {
			msg.Error(ctx, err.Error())
			return
		}
		msg.Ok(ctx, "ok")
		return
	}
	msg.Error(ctx, "invalid file type")
}
