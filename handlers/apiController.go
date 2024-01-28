package handlers

import (
	"bytes"
	"database/sql"
	"errors"
	"github.com/CCLooMi/sql-mak/mysql"
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"github.com/gin-gonic/gin"
	"github.com/robertkrimen/otto"
	"io"
	"net/http"
	"time"
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
	ID     *string `json:"id"`
	Script *string `json:"script"`
	Args   []any   `json:"args"`
}

func NewApiController(app *gin.Engine) *ApiController {
	ctrl := &ApiController{db: conf.Db, apiService: service.NewApiService(conf.Db)}
	group := app.Group("/api")
	hds := []middlewares.Auth{
		{Method: "POST", Group: "/api", Path: "/execute", Auth: "api.execute", Handler: ctrl.execute},
		{Method: "POST", Group: "/api", Path: "/executeById", Auth: "api.executeById", Handler: ctrl.executeById},
		{Method: "POST", Group: "/api", Path: "/byPage", Auth: "api.list", Handler: ctrl.byPage},
		{Method: "POST", Group: "/api", Path: "/saveUpdate", Auth: "api.saveUpdate", Handler: ctrl.saveUpdate},
		{Method: "POST", Group: "/api", Path: "/saveUpdates", Auth: "api.saveUpdates", Handler: ctrl.saveUpdates},
		{Method: "POST", Group: "/api", Path: "/delete", Auth: "api.delete", Handler: ctrl.delete},
	}
	for i, hd := range hds {
		middlewares.RegisterAuth(&hds[i])
		group.Handle(hd.Method, hd.Path, hd.Handler)
	}
	return ctrl
}

type mySQLStruct struct {
	SELECT        any
	SELECT_EXP    any
	SELECT_AS     any
	SELECT_SM_AS  any
	SELECT_EXP_AS any
	INSERT_INTO   any
	UPDATE        any
	DELETE        any
	TxExecute     any
}

var mysqlM = mySQLStruct{
	mysql.SELECT,
	mysql.SELECT_EXP,
	mysql.SELECT_AS,
	mysql.SELECT_SM_AS,
	mysql.SELECT_EXP_AS,
	mysql.INSERT_INTO,
	mysql.UPDATE,
	mysql.DELETE,
	mysql.TxExecute,
}

type expStruct struct {
	Now    any
	UUID   any
	Exp    any
	ExpStr any
}

var expM = expStruct{
	mak.Now,
	mak.UUID,
	mak.Exp,
	mak.ExpStr,
}
var halt = errors.New("Stahp")

func runUnsafe(unsafe string, timeout time.Duration, c *gin.Context, args []any) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if caught := recover(); caught != nil {
			if caught == halt {
				msg.Error(c, "Some code took to long! Stopping after: "+duration.String())
				return
			}
			// Something else happened, repanic!
			panic(caught)
		}
	}()
	vm := otto.New()
	vm.Interrupt = make(chan func(), 1)
	vm.Set("request", c)
	vm.Set("msgOk", func(data any) {
		msg.Ok(c, data)
	})
	vm.Set("msgError", func(err error) {
		msg.Error(c, err)
	})
	vm.Set("msgOks", func(data ...any) {
		msg.Oks(c, data...)
	})
	vm.Set("byPage", func(f func(sm *mak.SQLSM)) {
		middlewares.ByPage(c, func(page *middlewares.Page) (int64, any, error) {
			sm := mak.NewSQLSM()
			f(sm)
			sm.LIMIT(page.PageNumber*page.PageSize, page.PageSize)
			out := sm.Execute(conf.Db).GetResultAsMapList()
			if page.PageNumber == 0 {
				return sm.Execute(conf.Db).Count(), out, nil
			}
			return 0, out, nil
		})
	})
	vm.Set("db", conf.Db)
	vm.Set("rdb", conf.Rdb)
	vm.Set("cfg", conf.Cfg)
	vm.Set("sql", mysqlM)
	vm.Set("exp", expM)
	vm.Set("args", args)
	vm.Set("fetch", fetch)
	watchdogCleanup := make(chan struct{})
	defer close(watchdogCleanup)
	go func() {
		select {
		// Stop after timeout
		case <-time.After(timeout * time.Second):
			vm.Interrupt <- func() {
				panic(halt)
			}
		case <-watchdogCleanup:
		}
		close(vm.Interrupt)
	}()
	result, err := vm.Run(unsafe)
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
}
func fetch(url string, o ...interface{}) (map[string]interface{}, error) {
	var opts map[string]interface{}
	if len(o) > 0 {
		opts = o[0].(map[string]interface{})
	} else {
		opts = map[string]interface{}{}
	}
	method, methodExists := opts["method"].(string)
	if !methodExists || method == "" {
		method = "GET"
	}
	var body string
	if bodyInterface, bodyExists := opts["body"]; bodyExists {
		body = bodyInterface.(string)
	}
	req, err := http.NewRequest(method, url, bytes.NewBufferString(body))
	if err != nil {
		return nil, err
	}
	headers, headersExist := opts["headers"].(map[string]interface{})
	if headersExist {
		for k, v := range headers {
			req.Header.Set(k, v.(string))
		}
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	rspBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"Response":   string(rspBody),
		"Request":    resp.Request,
		"Status":     resp.Status,
		"StatusCode": resp.StatusCode,
		"Header":     resp.Header,
		"Cookies": func() []*http.Cookie {
			return resp.Cookies()
		},
	}, nil
}
func (ctrl *ApiController) execute(c *gin.Context) {
	var reqBody ReqBody
	if err := c.BindJSON(&reqBody); err != nil {
		msg.Error(c, err)
		return
	}
	runUnsafe(*reqBody.Script, time.Duration(10), c, reqBody.Args)
}
func (ctrl *ApiController) executeById(c *gin.Context) {
	var reqBody ReqBody
	if err := c.BindJSON(&reqBody); err != nil {
		msg.Error(c, err)
		return
	}

	api := &entity.Api{}
	ctrl.apiService.ById(reqBody.ID, api)
	if api.Id == nil {
		msg.Error(c, "api not found")
		return
	}
	runUnsafe(*api.Script, time.Duration(10), c, reqBody.Args)
}
func (ctrl *ApiController) byPage(c *gin.Context) {
	middlewares.ByPage(c, func(page *middlewares.Page) (int64, any, error) {
		return ctrl.apiService.ListByPage(page.PageNumber, page.PageSize, func(sm *mak.SQLSM) {
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
	userInfo, ok := c.Get("userInfo")
	if !ok {
		msg.Error(c, "userInfo not found")
		return
	}
	userId := userInfo.(*middlewares.UserInfo).User.Id
	api.UpdatedBy = userId
	if api.CreatedBy == nil {
		api.CreatedBy = userId
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
		if apis[i].CreatedBy == nil {
			apis[i].CreatedBy = userId
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
