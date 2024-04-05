package handlers

import (
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
	"time"
	"wios_server/conf"
	"wios_server/entity"
	"wios_server/handlers/beans"
	"wios_server/handlers/msg"
	"wios_server/middlewares"
	"wios_server/service"
	"wios_server/utils"
)

type UserController struct {
	userService *service.UserService
}

func NewUserController(app *gin.Engine) *UserController {
	ctrl := &UserController{userService: service.NewUserService(conf.Db)}
	group := app.Group("/user")
	hds := []middlewares.Auth{
		{Method: "GET", Group: "/user", Path: "/initRoot", Handler: ctrl.initRootUser},
		{Method: "POST", Group: "/user", Path: "/byPage", Auth: "user.list", Handler: ctrl.byPage},
		{Method: "POST", Group: "/user", Path: "/saveUpdate", Auth: "user.update", Handler: ctrl.saveUpdate},
		{Method: "POST", Group: "/user", Path: "/delete", Auth: "user.delete", Handler: ctrl.delete},
		{Method: "POST", Group: "/user", Path: "/login", Handler: ctrl.login},
		{Method: "GET", Group: "/user", Path: "/current", Auth: "#", Handler: ctrl.currentUser},
		{Method: "GET", Group: "/user", Path: "/logout", Auth: "#", Handler: ctrl.logout},
		{Method: "GET", Group: "/user", Path: "/menus", Auth: "#", Handler: ctrl.menus},
		{Method: "POST", Group: "/user", Path: "/roles", Auth: "#", Handler: ctrl.roles},
		{Method: "POST", Group: "/user", Path: "/addRole", Auth: "#", Handler: ctrl.addRole},
		{Method: "POST", Group: "/user", Path: "/removeRole", Auth: "#", Handler: ctrl.removeRole},
	}
	for i, hd := range hds {
		middlewares.RegisterAuth(&hds[i])
		group.Handle(hd.Method, hd.Path, hd.Handler)
	}
	return ctrl
}
func (ctrl *UserController) byPage(ctx *gin.Context) {
	middlewares.ByPage(ctx, func(page *middlewares.Page) (int64, any, error) {
		return ctrl.userService.ListByPage(page.PageNumber, page.PageSize, func(sm *mak.SQLSM) {
			sm.SELECT("*").FROM(entity.User{}, "u")
		})
	})
}

func (ctrl *UserController) saveUpdate(ctx *gin.Context) {
	var user entity.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	if user.Id == nil {
		if ctrl.userService.CheckExist(&entity.User{Username: user.Username}) {
			msg.Error(ctx, "username exists")
			return
		}
		user.Seed = utils.RandomBytes(8)
		user.Password = utils.SHA256(user.Username, user.Password, user.Seed)
	} else {
		oldUser, _ := ctrl.userService.FindById(user.Id)
		if oldUser.Id != nil {
			if user.Password == "" {
				user.Password = oldUser.Password
			}
			if user.Username == "" {
				user.Username = oldUser.Username
			}
			if oldUser.Password != user.Password || oldUser.Username != user.Username {
				user.Password = utils.SHA256(user.Username, user.Password, oldUser.Seed)
			}
		}
	}
	var rs = ctrl.userService.SaveUpdateWithFilter(&user,
		func(fieldName *string, columnName *string, v interface{}, im *mak.SQLIM) bool {
			if utils.IsBlank(v) {
				return false
			}
			return true
		})
	_, err := rs.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	userInfo := ctx.MustGet(middlewares.UserInfoKey).(*middlewares.UserInfo)
	if *userInfo.User.Id == *user.Id {
		userInfo.User = &user
		utils.SaveObjDataToRedis(userInfo.Id, userInfo, time.Hour*24)
	}
	user.Password = ""
	user.Seed = nil
	msg.Ok(ctx, &user)
}
func (ctrl *UserController) delete(ctx *gin.Context) {
	var user entity.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	var rs = ctrl.userService.DeleteUser(&user)
	affected, err := rs[1].RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	if affected > 0 {
		msg.Ok(ctx, &user)
		return
	}
	msg.Error(ctx, "delete failed")
}
func (ctrl *UserController) login(ctx *gin.Context) {
	var userInfo map[string]string
	if err := ctx.ShouldBindJSON(&userInfo); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	user, roles, pm := ctrl.userService.FindByUsernameAndPassword(userInfo["username"], userInfo["password"])
	if user == nil {
		msg.Error(ctx, "username or password error")
		return
	}

	CID, _ := ctx.Cookie(middlewares.UserSessionIDKey)
	if CID != "" {
		utils.DelFromRedis(CID)
	}
	CID = utils.GenerateRandomID()
	domain := utils.RemoveDomainPort(ctx.Request.Host)
	maxAge := 60 * 60 * 24
	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     middlewares.UserSessionIDKey,
		Value:    url.QueryEscape(CID),
		MaxAge:   maxAge,
		Path:     "/",
		Domain:   domain,
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
		HttpOnly: true,
	})
	infoMap := map[string]interface{}{
		"id":          CID,
		"user":        user,
		"roles":       roles,
		"permissions": pm,
	}
	err := utils.SaveObjDataToRedis(CID, infoMap, time.Hour*24)
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	user.Password = ""
	user.Seed = nil
	msg.Ok(ctx, user)
}

func (ctrl *UserController) currentUser(ctx *gin.Context) {
	userInfo := ctx.MustGet(middlewares.UserInfoKey).(*middlewares.UserInfo)
	userInfo.User.Password = ""
	userInfo.User.Seed = nil
	msg.Ok(ctx, userInfo)
}

func (ctrl *UserController) logout(ctx *gin.Context) {
	CID, err := ctx.Cookie(middlewares.UserSessionIDKey)
	if err != nil || CID == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "Unauthorized",
		})
		return
	}
	utils.DelFromRedis(CID)
	msg.Ok(ctx, nil)
}

func (ctrl *UserController) menus(ctx *gin.Context) {
	userInfo := ctx.MustGet("userInfo").(*middlewares.UserInfo)
	menus := ctrl.userService.FindMenusByUser(userInfo.User)
	msg.Ok(ctx, menus)
}

func (ctrl *UserController) roles(ctx *gin.Context) {
	var pageInfo beans.PageInfo
	if err := ctx.ShouldBindJSON(&pageInfo); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	data := ctrl.userService.FindRolesByUserId(pageInfo.Opts["userId"].(string), pageInfo.Page, pageInfo.PageSize, pageInfo.Opts["yes"].(bool))
	msg.Ok(ctx, data)
}

func (ctrl *UserController) addRole(ctx *gin.Context) {
	var roleUser entity.RoleUser
	if err := ctx.ShouldBindJSON(&roleUser); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	r := ctrl.userService.AddRole(&roleUser)
	rows, err := r.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	if rows > 0 {
		msg.Ok(ctx, &roleUser)
		return
	}
	msg.Error(ctx, "save failed")
}

func (ctrl *UserController) removeRole(ctx *gin.Context) {
	var roleUser entity.RoleUser
	if err := ctx.ShouldBindJSON(&roleUser); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	r := ctrl.userService.RemoveRole(&roleUser)
	rows, err := r.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	if rows > 0 {
		msg.Ok(ctx, &roleUser)
		return
	}
	msg.Error(ctx, "save failed")
}
func (ctrl *UserController) initRootUser(ctx *gin.Context) {
	user := entity.User{
		Username: "root",
		Nickname: "Super Admin",
		Password: "apple",
	}
	if user.Id == nil {
		if ctrl.userService.CheckExist(&entity.User{Username: user.Username}) {
			msg.Ok(ctx, "root user already exists")
			return
		}
	}
	if user.Seed == nil {
		user.Seed = utils.RandomBytes(8)
		user.Password = utils.SHA256(user.Username, user.Password, user.Seed)
	}
	var rs = ctrl.userService.SaveUpdate(&user)
	_, err := rs.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	msg.Ok(ctx, "init root user success")
}
