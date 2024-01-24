package handlers

import (
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"github.com/gin-gonic/gin"
	"wios_server/conf"
	"wios_server/entity"
	"wios_server/handlers/beans"
	"wios_server/handlers/msg"
	"wios_server/middlewares"
	"wios_server/service"
)

type RoleController struct {
	roleService *service.RoleService
}

func NewRoleController(app *gin.Engine) *RoleController {
	ctrl := &RoleController{roleService: service.NewRoleService(conf.Db)}
	group := app.Group("/role")
	hds := []middlewares.Auth{
		{Method: "POST", Group: "/role", Path: "/byPage", Auth: "role.list", Handler: ctrl.byPage},
		{Method: "POST", Group: "/role", Path: "/saveUpdate", Auth: "role.update", Handler: ctrl.saveUpdate},
		{Method: "POST", Group: "/role", Path: "/delete", Auth: "role.delete", Handler: ctrl.delete},
		{Method: "POST", Group: "/role", Path: "/menus", Auth: "role.menus", Handler: ctrl.menus},
		{Method: "POST", Group: "/role", Path: "/permissions", Auth: "role.permissions", Handler: ctrl.permissions},
		{Method: "POST", Group: "/role", Path: "/users", Auth: "role.users", Handler: ctrl.users},
		{Method: "POST", Group: "/role", Path: "/addUser", Auth: "role.addUser", Handler: ctrl.addUser},
		{Method: "POST", Group: "/role", Path: "/removeUser", Auth: "role.removeUser", Handler: ctrl.removeUser},
		{Method: "POST", Group: "/role", Path: "/addMenu", Auth: "role.addMenu", Handler: ctrl.addMenu},
		{Method: "POST", Group: "/role", Path: "/removeMenu", Auth: "role.removeMenu", Handler: ctrl.removeMenu},
		{Method: "POST", Group: "/role", Path: "/updateMenus", Auth: "role.updateMenus", Handler: ctrl.updateMenus},
	}
	for i, hd := range hds {
		middlewares.AuthMap[hd.Group+hd.Path] = &hds[i]
		group.Handle(hd.Method, hd.Path, hd.Handler)
	}
	return ctrl
}
func (ctrl *RoleController) byPage(ctx *gin.Context) {
	middlewares.ByPage(ctx, func(pageNumber int, pageSize int) (int64, any, error) {
		return ctrl.roleService.ListByPage(pageNumber, pageSize, func(sm *mak.SQLSM) {
			sm.SELECT("*").FROM(entity.Role{}, "r")
		})
	})
}

func (ctrl *RoleController) saveUpdate(ctx *gin.Context) {
	var role entity.Role
	if err := ctx.ShouldBindJSON(&role); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	var rs = ctrl.roleService.SaveUpdate(&role)
	affected, err := rs.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	if affected > 0 {
		msg.Ok(ctx, &role)
		return
	}
	msg.Error(ctx, "saveUpdate failed")
}

func (ctrl *RoleController) delete(ctx *gin.Context) {
	var role entity.Role
	if err := ctx.ShouldBindJSON(&role); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	var rs = ctrl.roleService.DeleteRole(&role)
	affected, err := rs[0].RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	if affected > 0 {
		msg.Ok(ctx, &role)
		return
	}
	msg.Error(ctx, "delete failed")
}

func (ctrl *RoleController) menus(ctx *gin.Context) {
	var role entity.Role
	if err := ctx.ShouldBindJSON(&role); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	menus := ctrl.roleService.FindMenusByRole(&role)
	msg.Ok(ctx, menus)
}

func (ctrl *RoleController) permissions(ctx *gin.Context) {
	var role entity.Role
	if err := ctx.ShouldBindJSON(&role); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	ps := make([]map[string]interface{}, 0)
	pMap := ctrl.roleService.FindPermissionsByRole(&role)
	for _, v := range middlewares.AuthMap {
		ps = append(ps, map[string]interface{}{
			"id":      v.GetId(),
			"method":  v.Method,
			"group":   v.Group,
			"path":    v.Path,
			"auth":    v.Auth,
			"checked": pMap[v.GetId()],
		})
	}
	msg.Ok(ctx, ps)
}
func (ctrl *RoleController) users(ctx *gin.Context) {
	var pageInfo beans.PageInfo
	if err := ctx.ShouldBindJSON(&pageInfo); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	data := ctrl.roleService.FindUsersByRoleId(pageInfo.Opts["roleId"].(string), pageInfo.Page, pageInfo.PageSize, pageInfo.Opts["yes"].(bool))
	msg.Ok(ctx, data)
}

func (ctrl *RoleController) addMenu(ctx *gin.Context) {
	var roleMenu entity.RoleMenu
	if err := ctx.ShouldBindJSON(&roleMenu); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	var rs = ctrl.roleService.AddMenu(&roleMenu)
	affected, err := rs.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	if affected > 0 {
		msg.Ok(ctx, &roleMenu)
		return
	}
	msg.Error(ctx, "save failed")
}

func (ctrl *RoleController) removeMenu(ctx *gin.Context) {
	var roleMenu entity.RoleMenu
	if err := ctx.ShouldBindJSON(&roleMenu); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	var rs = ctrl.roleService.RemoveMenu(&roleMenu)
	affected, err := rs.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	if affected > 0 {
		msg.Ok(ctx, &roleMenu)
		return
	}
	msg.Error(ctx, "save failed")
}

type UpdateMenus struct {
	Add []entity.RoleMenu `json:"add"`
	Del []interface{}     `json:"del"`
}

func (ctrl *RoleController) updateMenus(ctx *gin.Context) {
	var updateMenus UpdateMenus
	if err := ctx.ShouldBindJSON(&updateMenus); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	r := ctrl.roleService.UpdateMenus(updateMenus.Add, updateMenus.Del)
	rows1, err := r[0].RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	rows2, err := r[1].RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	msg.Ok(ctx, rows1+rows2)
}

func (ctrl *RoleController) addUser(ctx *gin.Context) {
	var roleUser entity.RoleUser
	if err := ctx.ShouldBindJSON(&roleUser); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	var rs = ctrl.roleService.AddUser(&roleUser)
	affected, err := rs.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	if affected > 0 {
		msg.Ok(ctx, &roleUser)
		return
	}
	msg.Error(ctx, "save failed")
}

func (ctrl *RoleController) removeUser(ctx *gin.Context) {
	var roleUser entity.RoleUser
	if err := ctx.ShouldBindJSON(&roleUser); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	var rs = ctrl.roleService.RemoveUser(&roleUser)
	affected, err := rs.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	if affected > 0 {
		msg.Ok(ctx, &roleUser)
		return
	}
	msg.Error(ctx, "save failed")
}
