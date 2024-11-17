package handlers

import (
	"database/sql"
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"github.com/gin-gonic/gin"
	"wios_server/entity"
	"wios_server/handlers/msg"
	"wios_server/js"
	"wios_server/middlewares"
	"wios_server/service"
	"wios_server/utils"
)

type ApiController struct {
	db         *sql.DB
	apiService *service.ApiService
}

func NewApiController(app *gin.Engine, db *sql.DB, ut *utils.Utils, ac *middlewares.AuthChecker) *ApiController {
	ctrl := &ApiController{db: db, apiService: service.NewApiService(db, ut)}
	group := app.Group("/api")
	hds := []middlewares.Auth{
		{Method: "GET", Group: "/api", Path: "/stopVmById", Auth: "api.stopVmById", Handler: ctrl.stopVmById},
		{Method: "POST", Group: "/api", Path: "/vms", Auth: "api.vms", Handler: ctrl.vms},
		{Method: "POST", Group: "/api", Path: "/execute", Auth: "api.execute", Handler: ctrl.execute},
		{Method: "POST", Group: "/api", Path: "/executeById", Auth: "#", Handler: ctrl.executeById, AuthCheck: ac.ScriptApiAuthCheck},
		{Method: "POST", Group: "/api", Path: "/byPage", Auth: "api.list", Handler: ctrl.byPage},
		{Method: "POST", Group: "/api", Path: "/saveUpdate", Auth: "api.saveUpdate", Handler: ctrl.saveUpdate},
		{Method: "POST", Group: "/api", Path: "/saveUpdates", Auth: "api.saveUpdates", Handler: ctrl.saveUpdates},
		{Method: "POST", Group: "/api", Path: "/delete", Auth: "api.delete", Handler: ctrl.delete},
		{Method: "GET", Group: "/api", Path: "/backup", Auth: "api.backup", Handler: ctrl.backup},
	}
	for i, hd := range hds {
		middlewares.RegisterAuth(&hds[i])
		group.Handle(hd.Method, hd.Path, hd.Handler)
	}
	return ctrl
}

func (ctrl *ApiController) runUnsafe(unsafe string, title *string, c *gin.Context, args interface{}, reqBody map[string]interface{}) {
	ui, ok := c.Get("userInfo")
	if !ok {
		msg.Error(c, "userInfo not found")
		return
	}
	userInfo, ok := ui.(*middlewares.UserInfo)
	if !ok {
		msg.Error(c, "userInfo not found")
		return
	}
	var vm = js.NewVm(title, userInfo.User)
	//vm.Set("ctx", c)
	vm.Set("reqBody", reqBody)
	vm.Set("userInfo", userInfo)
	vm.Set("args", args)
	vm.Set("msg", js.NewMsgUtil(c))
	vm.Set("byPage", func(f func(sm *mak.SQLSM, opts interface{})) {
		middlewares.ByPageMap(reqBody, c, func(page *middlewares.Page) (int64, any, error) {
			if page.PageNumber < 0 {
				page.PageNumber = 0
			} else {
				page.PageNumber -= 1
			}
			sm := mak.NewSQLSM()
			f(sm, page.Opts)
			sm.LIMIT(page.PageNumber*page.PageSize, page.PageSize)
			out := sm.Execute(ctrl.db).GetResultAsMapList()
			if page.PageNumber == 0 {
				return sm.Execute(ctrl.db).Count(), out, nil
			}
			return -1, out, nil
		})
	})
	result, err := vm.Execute(unsafe)
	if err != nil {
		msg.Error(c, err.Error())
		return
	}
	if result.IsBoolean() {
		b, _ := result.ToBoolean()
		if !b {
			return
		}
	}
	msg.Ok(c, result)
	return
}
func (ctrl *ApiController) vms(c *gin.Context) {
	data := js.VMList()
	msg.Ok(c, map[string]interface{}{
		"data":  data,
		"total": len(data),
	})
}
func (ctrl *ApiController) stopVmById(c *gin.Context) {
	vmId := c.Query("id")
	msg.Ok(c, js.StopVM(vmId))
}
func (ctrl *ApiController) execute(c *gin.Context) {
	var reqBody map[string]interface{}
	if err := c.BindJSON(&reqBody); err != nil {
		msg.Error(c, err)
		return
	}
	id, ok := reqBody["id"].(string)
	if !ok {
		id = utils.UUID()
	}
	script, ok := reqBody["script"].(string)
	if !ok {
		msg.Ok(c, "")
		return
	}
	args, ok := reqBody["args"]
	if !ok {
		args = map[string]interface{}{}
	}
	ctrl.runUnsafe(script, &id, c, args, reqBody)
}
func (ctrl *ApiController) executeById(c *gin.Context) {
	apiInfo := c.MustGet(middlewares.ApiInfoKey).(*middlewares.ApiInfo)
	api := apiInfo.Api
	ctrl.runUnsafe(*api.Script, api.Desc, c, apiInfo.Args, apiInfo.ReqBody)
}
func (ctrl *ApiController) byPage(c *gin.Context) {
	middlewares.ByPage(c, func(page *middlewares.Page) (int64, any, error) {
		return ctrl.apiService.ListByPage(page.PageNumber, page.PageSize, func(sm *mak.SQLSM) {
			sm.SELECT("*").FROM(entity.Api{}, "a")
			q := page.Opts["q"]
			if q != nil && q != "" {
				lik := "%" + q.(string) + "%"
				sm.AND("(a.id = ? OR a.desc LIKE ? OR a.category LIKE ?)", q, lik, lik)
			}
			sm.ORDER_BY("a.updated_at DESC", "a.inserted_at DESC", "a.id")
		})
	})
}

func (ctrl *ApiController) saveUpdate(c *gin.Context) {
	var api entity.Api
	if err := c.ShouldBindJSON(&api); err != nil {
		msg.Error(c, err)
		return
	}
	userInfo := c.MustGet(middlewares.UserInfoKey).(*middlewares.UserInfo)
	api.UpdatedBy = userInfo.User.Id
	if api.InsertedBy == nil {
		api.InsertedBy = userInfo.User.Id
	}
	if api.UpdatedAt != nil {
		api.UpdatedAt = nil
	}
	var rs = ctrl.apiService.SaveUpdate(&api)
	_, err := rs.RowsAffected()
	if err != nil {
		msg.Error(c, err)
		return
	}
	msg.Ok(c, &api)
}
func (ctrl *ApiController) saveUpdates(c *gin.Context) {
	var apis []entity.Api
	if err := c.ShouldBindJSON(&apis); err != nil {
		msg.Error(c, err)
		return
	}
	userInfo, ok := c.Get("userInfo")
	if !ok {
		msg.Error(c, "userInfo not found")
		return
	}
	userId := userInfo.(*middlewares.UserInfo).User.Id
	for i := 0; i < len(apis); i++ {
		apis[i].UpdatedBy = userId
		if apis[i].InsertedBy == nil {
			apis[i].InsertedBy = userId
		}
	}
	var rs = ctrl.apiService.SaveUpdates(apis)
	for _, r := range rs {
		_, err := r.RowsAffected()
		if err != nil {
			msg.Error(c, err)
			return
		}
	}
	msg.Ok(c, &apis)
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

func (ctrl *ApiController) backup(c *gin.Context) {
	err := ctrl.apiService.Backup()
	if err != nil {
		msg.Error(c, err)
		return
	}
	msg.Ok(c, "ok")
}
