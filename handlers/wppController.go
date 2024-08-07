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

type WppController struct {
	wppService         *service.WppService
	releaseNoteService *service.ReleaseNoteService
	storeUserService   *service.StoreUserService
	ut                 *utils.Utils
}

func NewWppController(app *gin.Engine, db *sql.DB, ut *utils.Utils, ac *middlewares.AuthChecker) *WppController {
	ctrl := &WppController{
		wppService:         service.NewWppService(db),
		releaseNoteService: service.NewReleaseNoteService(db),
		storeUserService:   service.NewStoreUserService(db),
		ut:                 ut,
	}
	group := app.Group("/wpp")
	hds := []middlewares.Auth{
		{Method: "POST", Group: "/wpp", Path: "/topList", Handler: ctrl.topWpps},
		{Method: "POST", Group: "/wpp", Path: "/byPage", Auth: "wpp.byPage", Handler: ctrl.byPage},
		{Method: "POST", Group: "/wpp", Path: "/publish", Auth: "#", Handler: ctrl.publish, AuthCheck: ac.StoreAuthCheck},
		{Method: "POST", Group: "/wpp", Path: "/info", Handler: ctrl.getWppInfo},
	}
	for i, hd := range hds {
		middlewares.RegisterAuth(&hds[i])
		group.Handle(hd.Method, hd.Path, hd.Handler)
	}
	return ctrl
}
func (ctrl *WppController) byPage(ctx *gin.Context) {
	middlewares.ByPage(ctx, func(page *middlewares.Page) (int64, any, error) {
		return ctrl.wppService.ListByPage(page.PageNumber, page.PageSize, func(sm *mak.SQLSM) {
			sm.SELECT("*").FROM(entity.Wpp{}, "w")
		})
	})
}
func (ctrl *WppController) publish(ctx *gin.Context) {
	var m map[string]string
	if err := ctx.ShouldBindJSON(&m); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	version := m["version"]
	wppId := m["wppId"]
	fileId := m["fileId"]
	name := m["name"]
	releaseNote := m["releaseNote"]
	if version == "" || wppId == "" || fileId == "" || name == "" || releaseNote == "" {
		msg.Error(ctx, "param error")
		return
	}
	if isLatestV, latestV := ctrl.wppService.IsLatestVersion(&wppId, &version); !isLatestV {
		msg.Error(ctx, version+" is not latest version,current latest version is "+*latestV)
		return
	}
	userInfo := ctx.MustGet(middlewares.StoreUserInfoKey).(*middlewares.StoreUserInfo)
	user := userInfo.User
	manifest := m["manifest"]
	wpp := entity.Wpp{
		Name:          &name,
		Manifest:      &manifest,
		LatestVersion: &version,
		DeveloperId:   user.Id,
		FileId:        &fileId,
	}
	wpp.Id = &wppId
	oldWpp := ctrl.wppService.FindById(wpp.Id)
	if oldWpp.Id != nil {
		if *oldWpp.DeveloperId != *user.Id {
			msg.Error(ctx, "you are not the owner")
			return
		}
	}
	r := ctrl.wppService.SaveUpdate(&wpp)
	_, err := r.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	rNote := entity.ReleaseNote{
		WppId:       &wppId,
		Version:     &version,
		ReleaseNote: &releaseNote,
		DeveloperId: user.Id,
		FileId:      &fileId,
	}
	r = ctrl.releaseNoteService.SaveUpdate(&rNote)
	_, err = r.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	msg.Ok(ctx, "ok")
}
func (ctrl *WppController) topWpps(ctx *gin.Context) {
	m := make(map[string]interface{})
	if err := ctx.ShouldBindJSON(&m); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	t, ok := m["t"].(float64)
	if !ok {
		t = 0
	}
	limit, ok := m["limit"].(float64)
	if !ok {
		limit = 30
	}
	q, ok := m["q"].(string)
	if !ok {
		q = ""
	}
	wpps := ctrl.wppService.TopWpps(q, int(t), int(limit))
	msg.Ok(ctx, wpps)
}
func (ctrl *WppController) getWppInfo(ctx *gin.Context) {
	m := []string{}
	if err := ctx.ShouldBindJSON(&m); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	if len(m) == 0 {
		msg.Ok(ctx, m)
		return
	}
	if len(m) > 100 {
		msg.Error(ctx, "too many ids")
		return
	}
	storeUserInfo := middlewares.GetStoreUserInfo(ctx, ctrl.ut)
	var result = make([]map[string]interface{}, 0)
	for _, wppId := range m {
		wpp := ctrl.wppService.FindById(&wppId)
		if wpp.Id == nil {
			result = append(result, nil)
			continue
		}
		devUser, _ := ctrl.storeUserService.FindById(wpp.DeveloperId)
		var owner, isLogin = false, false
		if storeUserInfo != nil {
			owner = ctrl.storeUserService.Owner(storeUserInfo.User.Id, wpp.Id)
			isLogin = true
		}
		result = append(result, map[string]interface{}{
			"manifest":  wpp.Manifest,
			"fid":       wpp.FileId,
			"latestV":   wpp.LatestVersion,
			"owner":     owner,
			"isLogin":   isLogin,
			"devId":     devUser.Id,
			"devName":   devUser.Nickname,
			"devAvatar": devUser.Avatar,
			"devEmail":  devUser.Email,
		})
	}
	msg.Ok(ctx, result)
}
