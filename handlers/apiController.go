package handlers

import (
	"database/sql"
	"errors"
	"github.com/CCLooMi/sql-mak/mysql"
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"github.com/gin-gonic/gin"
	"github.com/robertkrimen/otto"
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
		{Method: "POST", Group: "/api", Path: "/saveUpdate", Auth: "api.update", Handler: ctrl.saveUpdate},
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

type makStruct struct {
	Now    any
	UUID   any
	Exp    any
	ExpStr any
}

var makM = makStruct{
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
		middlewares.ByPage(c, func(pageNumber int, pageSize int) (int64, any, error) {
			if pageNumber <= 0 {
				pageNumber = 0
			} else {
				pageNumber = pageNumber - 1
			}
			if pageSize <= 0 {
				pageSize = 20
			}
			sm := mak.NewSQLSM()
			f(sm)
			sm.LIMIT(pageNumber*pageSize, pageSize)
			out := sm.Execute(conf.Db).GetResultAsMapList()
			if pageNumber == 0 {
				return sm.Execute(conf.Db).Count(), out, nil
			}
			return 0, out, nil
		})
	})
	vm.Set("db", conf.Db)
	vm.Set("rdb", conf.Rdb)
	vm.Set("cfg", conf.Cfg)
	vm.Set("sql", mysqlM)
	vm.Set("mak", makM)
	vm.Set("args", args)
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
