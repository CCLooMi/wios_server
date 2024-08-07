package handlers

import (
	"database/sql"
	"encoding/json"
	"wios_server/entity"
	"wios_server/handlers/msg"
	"wios_server/middlewares"
	"wios_server/service"

	"github.com/CCLooMi/sql-mak/mysql/mak"
	"github.com/gin-gonic/gin"
)

type MenuController struct {
	menuService *service.MenuService
}

func NewMenuController(app *gin.Engine, db *sql.DB) *MenuController {
	ctrl := &MenuController{menuService: service.NewMenuService(db)}
	group := app.Group("/menu")
	hds := []middlewares.Auth{
		{Method: "POST", Group: "/menu", Path: "/byPage", Auth: "menu.list", Handler: ctrl.byPage},
		{Method: "POST", Group: "/menu", Path: "/saveUpdate", Auth: "menu.update", Handler: ctrl.saveUpdate},
		{Method: "POST", Group: "/menu", Path: "/delete", Auth: "menu.delete", Handler: ctrl.delete},
		{Method: "GET", Group: "/menu", Path: "/init", Auth: "menu.init", Handler: ctrl.initMenus},
	}
	for i, hd := range hds {
		middlewares.RegisterAuth(&hds[i])
		group.Handle(hd.Method, hd.Path, hd.Handler)
	}
	return ctrl
}

func (ctrl *MenuController) byPage(ctx *gin.Context) {
	middlewares.ByPage(ctx, func(page *middlewares.Page) (int64, any, error) {
		return ctrl.menuService.ListByPage(page.PageNumber, page.PageSize, func(sm *mak.SQLSM) {
			sm.SELECT("*").FROM(entity.Menu{}, "m")
		})
	})
}

func (ctrl *MenuController) saveUpdate(ctx *gin.Context) {
	var menu entity.Menu
	if err := ctx.ShouldBindJSON(&menu); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	userInfo := ctx.MustGet(middlewares.UserInfoKey).(*middlewares.UserInfo)
	menu.UpdatedBy = userInfo.User.Id
	if menu.InsertedBy == nil {
		menu.InsertedBy = userInfo.User.Id
	}
	if menu.UpdatedAt != nil {
		menu.UpdatedAt = nil
	}
	var rs = ctrl.menuService.SaveUpdate(&menu)
	_, err := rs.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	msg.Ok(ctx, &menu)
}

func (ctrl *MenuController) delete(ctx *gin.Context) {
	var menu entity.Menu
	if err := ctx.ShouldBindJSON(&menu); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	var rs = ctrl.menuService.DeleteMenu(&menu)
	affected, err := (rs[1]).RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	if affected > 0 {
		msg.Ok(ctx, &menu)
		return
	}
	msg.Error(ctx, "delete failed")
}

func (ctrl *MenuController) initMenus(ctx *gin.Context) {
	menuMap := []map[string]interface{}{
		{"id": "237372500b86260b748e95143587c991", "rootId": "2a9533d1aba99986babeece48ef2c1bc", "pid": "2a9533d1aba99986babeece48ef2c1bc", "idx": 0, "name": "Menus", "href": "main.menus"},
		{"id": "ed709984b8d011ee82370242ac120002", "rootId": "2a9533d1aba99986babeece48ef2c1bc", "pid": "2a9533d1aba99986babeece48ef2c1bc", "idx": 0, "name": "Apis", "href": "main.apis"},
		{"id": "0017df5a91bb2cadc5fcc22e0a360b76", "rootId": "2a9533d1aba99986babeece48ef2c1bc", "pid": "2a9533d1aba99986babeece48ef2c1bc", "idx": 0, "name": "Configs", "href": "main.configs"},
		{"id": "1609474673d88caaa045ebaa8d9273c6", "rootId": "2a9533d1aba99986babeece48ef2c1bc", "pid": "2a9533d1aba99986babeece48ef2c1bc", "idx": 0, "name": "Uploads", "href": "main.uploads"},
		{"id": "2a9533d1aba99986babeece48ef2c1bc", "rootId": "2a9533d1aba99986babeece48ef2c1bc", "pid": "#", "idx": 0, "name": "System", "href": ""},
		{"id": "a658e46f2fe2699846bcf89053ae4001", "rootId": "a658e46f2fe2699846bcf89053ae4001", "pid": "#", "idx": 0, "name": "Security", "href": ""},
		{"id": "f687ac08d79f2d066dd0d2d6058f7f01", "rootId": "a658e46f2fe2699846bcf89053ae4001", "pid": "a658e46f2fe2699846bcf89053ae4001", "idx": 0, "name": "Users", "href": "main.users"},
		{"id": "f6b6af3a67dea5704da2a1150033063d", "rootId": "a658e46f2fe2699846bcf89053ae4001", "pid": "a658e46f2fe2699846bcf89053ae4001", "idx": 0, "name": "Roles", "href": "main.roles"},
	}
	menus := make([]entity.Menu, 0)
	jsonStr, err := json.Marshal(menuMap)
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	err = json.Unmarshal(jsonStr, &menus)
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	interfaceMenus := make([]interface{}, len(menus))
	userInfo := ctx.MustGet(middlewares.UserInfoKey).(*middlewares.UserInfo)
	uid := userInfo.User.Id
	for i, menu := range menus {
		menu.InsertedBy = uid
		menu.UpdatedBy = uid
		interfaceMenus[i] = menu
	}
	ctrl.menuService.BatchSaveUpdate(interfaceMenus...)
}
