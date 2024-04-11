package handlers

import (
	"crypto/sha1"
	"encoding/hex"
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
		{Method: "POST", Group: "/role", Path: "/updatePermissions", Auth: "role.updatePermissions", Handler: ctrl.updatePermissions},
	}
	for i, hd := range hds {
		middlewares.RegisterAuth(&hds[i])
		group.Handle(hd.Method, hd.Path, hd.Handler)
	}
	return ctrl
}
func (ctrl *RoleController) byPage(ctx *gin.Context) {
	middlewares.ByPage(ctx, func(page *middlewares.Page) (int64, any, error) {
		return ctrl.roleService.ListByPage(page.PageNumber, page.PageSize, func(sm *mak.SQLSM) {
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
	userInfo := ctx.MustGet(middlewares.UserInfoKey).(*middlewares.UserInfo)
	role.UpdatedBy = userInfo.User.Id
	if role.InsertedBy == nil {
		role.InsertedBy = userInfo.User.Id
	}
	if role.UpdatedAt != nil {
		role.UpdatedAt = nil
	}
	var rs = ctrl.roleService.SaveUpdate(&role)
	_, err := rs.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	msg.Ok(ctx, &role)
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
	gMap := make(map[string]map[string]interface{})
	for _, v := range middlewares.AuthList {
		group := gMap[v.Group]
		if group == nil {
			hash := sha1.Sum([]byte(v.Method + v.Group))
			id := hex.EncodeToString(hash[:])
			checked := ""
			if pMap[id] {
				checked = "on"
			}
			gMap[v.Group] = map[string]interface{}{
				"id":      id,
				"name":    v.Group,
				"checked": checked,
			}
			group = gMap[v.Group]
			ps = append(ps, group)
		}
		checked := ""
		if pMap[v.GetId()] {
			checked = "on"
		}
		ps = append(ps, map[string]interface{}{
			"id":      v.GetId(),
			"pid":     group["id"],
			"name":    v.Path,
			"auth":    v.Auth,
			"checked": checked,
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
	var c int64 = 0
	for _, v := range r {
		cc, err := v.RowsAffected()
		if err != nil {
			msg.Error(ctx, err.Error())
			return
		}
		c += cc
	}
	msg.Ok(ctx, c)
}

type UpdatePermissions struct {
	Add []entity.RolePermission `json:"add"`
	Del []interface{}           `json:"del"`
}

func (ctrl *RoleController) updatePermissions(ctx *gin.Context) {
	var updatePermissions UpdatePermissions
	if err := ctx.ShouldBindJSON(&updatePermissions); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	r := ctrl.roleService.UpdatePermissions(updatePermissions.Add, updatePermissions.Del)
	var c int64 = 0
	for _, v := range r {
		cc, err := v.RowsAffected()
		if err != nil {
			msg.Error(ctx, err.Error())
			return
		}
		c += cc
	}
	msg.Ok(ctx, c)

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
