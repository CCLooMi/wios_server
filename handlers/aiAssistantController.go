package handlers

import (
	"database/sql"
	"errors"
	"github.com/CCLooMi/sql-mak/mysql"
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"github.com/gin-gonic/gin"
	"log"
	"time"
	"wios_server/entity"
	"wios_server/handlers/msg"
	"wios_server/js"
	"wios_server/middlewares"
	"wios_server/service"
	"wios_server/utils"
)

type AiAssistantController struct {
	db                   *sql.DB
	apiService           *service.ApiService
	aiAssistantService   *service.AiAssistantService
	aiChatHistoryService *service.AiChatHistoryService
}

func NewAiAssistantController(app *gin.Engine, db *sql.DB, ut *utils.Utils) *AiAssistantController {
	ctrl := &AiAssistantController{db: db,
		apiService:           service.NewApiService(db, ut),
		aiAssistantService:   service.NewAiAssistantService(db, ut),
		aiChatHistoryService: service.NewAiChatHistoryService(db, ut),
	}
	group := app.Group("/aiAssistant")
	hds := []middlewares.Auth{
		{Method: "POST", Group: "/aiAssistant", Path: "/byPage", Auth: "aiAssistant.byPage", Handler: ctrl.byPage},
		{Method: "POST", Group: "/aiAssistant", Path: "/saveUpdate", Auth: "aiAssistant.saveUpdate", Handler: ctrl.saveUpdate},
		{Method: "POST", Group: "/aiAssistant", Path: "/saveUpdates", Auth: "aiAssistant.saveUpdates", Handler: ctrl.saveUpdates},
		{Method: "POST", Group: "/aiAssistant", Path: "/delete", Auth: "aiAssistant.delete", Handler: ctrl.delete},
		{Method: "POST", Group: "/aiAssistant", Path: "/chatHistory", Auth: "aiAssistant.chatHistory", Handler: ctrl.chatHistory},
		{Method: "POST", Group: "/aiAssistant", Path: "/run", Auth: "aiAssistant.run", Handler: ctrl.run},
	}
	for i, hd := range hds {
		middlewares.RegisterAuth(&hds[i])
		group.Handle(hd.Method, hd.Path, hd.Handler)
	}
	ctrl.startTask()
	return ctrl
}

func (ctrl *AiAssistantController) byPage(c *gin.Context) {
	middlewares.ByPage(c, func(page *middlewares.Page) (int64, any, error) {
		return ctrl.aiAssistantService.ListByPage(page.PageNumber, page.PageSize, func(sm *mak.SQLSM) {
			sm.SELECT("*").FROM(entity.AiAssistant{}, "a")
			q := page.Opts["q"]
			if q != nil && q != "" {
				lik := "%" + q.(string) + "%"
				sm.AND("(a.id = ? OR a.name = ? OR a.name like ?)", q, q, lik)
			}
			sm.ORDER_BY("a.updated_at DESC", "a.inserted_at DESC", "a.id")
		})
	})
}
func (ctrl *AiAssistantController) saveUpdate(c *gin.Context) {
	var aiAssistant entity.AiAssistant
	if err := c.ShouldBindJSON(&aiAssistant); err != nil {
		msg.Error(c, err.Error())
		return
	}
	if aiAssistant.UpdatedAt != nil {
		aiAssistant.UpdatedAt = nil
	}
	var rs = ctrl.aiAssistantService.SaveUpdate(&aiAssistant)
	_, err := rs.RowsAffected()
	if err != nil {
		msg.Error(c, err.Error())
		return
	}
	msg.Ok(c, &aiAssistant)
}
func (ctrl *AiAssistantController) saveUpdates(c *gin.Context) {
	var aiAssistants []entity.AiAssistant
	if err := c.ShouldBindJSON(&aiAssistants); err != nil {
		msg.Error(c, err.Error())
		return
	}
	for i := 0; i < len(aiAssistants); i++ {
		if aiAssistants[i].UpdatedAt != nil {
			aiAssistants[i].UpdatedAt = nil
		}
	}
	var rs = ctrl.aiAssistantService.SaveUpdates(aiAssistants)
	for _, r := range rs {
		_, err := r.RowsAffected()
		if err != nil {
			msg.Error(c, err)
			return
		}
	}
	msg.Ok(c, &aiAssistants)
}
func (ctrl *AiAssistantController) delete(c *gin.Context) {
	var aiAssistant entity.AiAssistant
	if err := c.ShouldBindJSON(&aiAssistant); err != nil {
		msg.Error(c, err.Error())
		return
	}
	var rs = ctrl.aiAssistantService.Delete(&aiAssistant)
	affected, err := rs.RowsAffected()
	if err != nil {
		msg.Error(c, err.Error())
		return
	}
	if affected > 0 {
		msg.Ok(c, &aiAssistant)
		return
	}
	msg.Error(c, "delete failed")
}
func (ctrl *AiAssistantController) run(c *gin.Context) {
	var reqBody map[string]interface{}
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		msg.Error(c, err.Error())
		return
	}
	aid, ok := reqBody["aid"].(string)
	if !ok || aid == "" {
		msg.Error(c, "invalid assistant id")
		return
	}

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
	err := ctrl.runAiAssistantById(aid, userInfo)
	if err != nil {
		msg.Error(c, err.Error())
		return
	}
	msg.Ok(c, "ok")
	return
}
func (ctrl *AiAssistantController) chatHistory(c *gin.Context) {
	middlewares.ByPage(c, func(page *middlewares.Page) (int64, any, error) {
		return ctrl.aiChatHistoryService.ListByPage(page.PageNumber, page.PageSize, func(sm *mak.SQLSM) {
			sm.SELECT("*").FROM(entity.AiChatHistory{}, "a")
			q := page.Opts["q"]
			if q != nil && q != "" {
				lik := "%" + q.(string) + "%"
				sm.AND("(a.id = ? a.assistantId = ? OR a.sessionId = ? OR a.msgId = ? OR a.role = ? OR a.content like ?)", q, q, q, q, q, lik)
			}
			sm.ORDER_BY("a.updated_at DESC", "a.inserted_at DESC", "a.id")
		})
	})
}

func (ctrl *AiAssistantController) startTask() {
	flag := func(flagId string) int64 {
		um := mysql.UPDATE(entity.AiAssistant{}, "a").
			SET("a.flagId = ?", flagId).
			SET("a.flagExp = (NOW(6)+?)", 10).
			WHERE("a.bootType = 1").
			AND("(a.status = 'stopped' OR a.status IS NULL)").
			AND("(a.flagExp < NOW(6) OR a.flagExp IS NULL)").
			LIMIT(3)
		um.LOGSQL(false)
		r := ctrl.aiAssistantService.ExecuteUm(um).Update()
		n, err := r.RowsAffected()
		if err != nil {
			log.Println(err)
		}
		return n
	}
	go func() {
		flagId := utils.UUID()
		userInfo := &middlewares.UserInfo{
			User: &entity.User{
				Nickname: "Ai Assistant Task",
			},
		}
		sm := mysql.SELECT("a.id").
			FROM(entity.AiAssistant{}, "a").
			WHERE("a.flagId =?", flagId).
			AND("a.bootType = 1").
			AND("(a.status = 'stopped' OR a.status IS NULL)").
			LIMIT(3)
		sm.LOGSQL(false)
		for {
			n := flag(flagId)
			if n == 0 {
				time.Sleep(10 * time.Second)
				continue
			}
			aiAssistants := ctrl.aiAssistantService.ExecuteSM(sm).GetResultAsList()
			for _, aiAssistant := range aiAssistants {
				aid := aiAssistant.(**string)
				ctrl.runAiAssistantById(**aid, userInfo)
			}
			time.Sleep(10 * time.Second)
		}
	}()
	log.Printf("ai assistant task started")
}
func (ctrl *AiAssistantController) runAiAssistantById(aid string, userInfo *middlewares.UserInfo) error {
	//recover
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	aiAssistant := &entity.AiAssistant{}
	ctrl.aiAssistantService.ById(aid, aiAssistant)
	if aiAssistant.Id == nil {
		return errors.New("invalid assistant id")
	}
	var api = &entity.Api{}
	ctrl.apiService.ById(*aiAssistant.ScriptId, api)
	if api.Id == nil {
		return errors.New("invalid script id")
	}
	var vm = js.NewVm(aiAssistant.Name, userInfo.User)
	vm.Set("userInfo", userInfo)
	vm.Set("aid", aid)
	vm.Set("aiAssistant", aiAssistant)
	vm.Finally(func() {
		ctrl.aiAssistantService.SetStatus(aiAssistant.Id, "stopped")
	})
	ctrl.aiAssistantService.SetStatus(aiAssistant.Id, "running")
	go vm.Execute(*api.Script)
	return nil
}
