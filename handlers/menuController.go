package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"wios_server/entity"
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
	group.POST("/listByPage", ctrl.byPage)
	group.POST("/saveUpdate", ctrl.saveUpdate)
	return ctrl
}

func (ctrl *MenuController) byPage(ctx *gin.Context) {
	// 获取请求参数
	pageNumber, _ := strconv.Atoi(ctx.Query("pageNumber"))
	pageSize, _ := strconv.Atoi(ctx.Query("pageSize"))
	count, menus, err := ctrl.menuService.ListByPage(pageNumber, pageSize, func(sm *mak.SQLSM) {
		sm.SELECT("*").FROM(entity.Menu{}, "m")
	})
	if err != nil {
		panic(err)
	}
	result := map[string]interface{}{
		"count": count,
		"data":  menus,
	}
	ctx.JSON(http.StatusOK, result)
}

func (ctrl *MenuController) saveUpdate(ctx *gin.Context) {
	//获取post请求的json对象
	var menu entity.Menu
	if err := ctx.ShouldBindJSON(&menu); err != nil {
		panic(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var rs = ctrl.menuService.SaveUpdate(&menu)
	ctx.JSON(http.StatusOK, rs)
}
