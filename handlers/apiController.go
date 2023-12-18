package handlers

import (
	"database/sql"
	"fmt"
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"github.com/gin-gonic/gin"
	"wios_server/conf"
	"wios_server/entity"
	"wios_server/handlers/msg"
	"wios_server/middlewares"
	"wios_server/service"
)

type ApiController struct {
	db         *sql.DB
	apiService *service.ApiService
}
type ReqBody struct {
	ID   string `json:"id"`
	Args []any  `json:"args"`
}

func NewApiController(app *gin.Engine) *ApiController {
	ctrl := &ApiController{db: conf.Db, apiService: service.NewApiService(conf.Db)}
	group := app.Group("/api")
	hds := []middlewares.Auth{
		{Method: "POST", Group: "/api", Path: "/execute", Auth: "api.execute", Handler: ctrl.execute},
		{Method: "POST", Group: "/api", Path: "/byPage", Auth: "api.list", Handler: ctrl.byPage},
		{Method: "POST", Group: "/api", Path: "/saveUpdate", Auth: "api.update", Handler: ctrl.saveUpdate},
		{Method: "POST", Group: "/api", Path: "/delete", Auth: "api.delete", Handler: ctrl.delete},
	}
	for _, hd := range hds {
		middlewares.AuthMap[hd.Group+hd.Path] = &hd
		group.Handle(hd.Method, hd.Path, hd.Handler)
	}
	return ctrl
}

func (ctrl *ApiController) execute(c *gin.Context) {
	var reqBody ReqBody
	if err := c.BindJSON(&reqBody); err != nil {
		msg.Error(c, err)
		return
	}

	api := &entity.Api{}
	ctrl.apiService.ById(reqBody.ID, api)

	fmt.Print(api)
}

func (ctrl *ApiController) byPage(c *gin.Context) {
	middlewares.ByPage(c, func(pageNumber int, pageSize int) (int64, any, error) {
		return ctrl.apiService.ListByPage(pageNumber, pageSize, func(sm *mak.SQLSM) {
			sm.SELECT("*").FROM(entity.Api{}, "a")
		})
	})
}

func (ctrl *ApiController) saveUpdate(c *gin.Context) {
	var api entity.Api
	if err := c.ShouldBindJSON(&api); err != nil {
		msg.Error(c, err)
		return
	}
	var rs = ctrl.apiService.SaveUpdate(&api)
	affected, err := rs.RowsAffected()
	if err != nil {
		msg.Error(c, err)
		return
	}
	if affected > 0 {
		msg.Ok(c, &api)
		return
	}
	msg.Error(c, "save failed")
}

func (ctrl *ApiController) delete(c *gin.Context) {
	var api entity.Api
	if err := c.ShouldBindJSON(&api); err != nil {
		msg.Error(c, err)
		return
	}
	var rs = ctrl.apiService.Delete(&api)
	affected, err := rs.RowsAffected()
	if err != nil {
		msg.Error(c, err)
		return
	}
	if affected > 0 {
		msg.Ok(c, &api)
		return
	}
	msg.Error(c, "delete failed")
}
