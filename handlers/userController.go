package handlers

import (
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
	"wios_server/conf"
	"wios_server/entity"
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
		{Method: "POST", Group: "/user", Path: "/byPage", Auth: "user.list", Handler: ctrl.byPage},
		{Method: "POST", Group: "/user", Path: "/saveUpdate", Auth: "user.update", Handler: ctrl.saveUpdate},
		{Method: "POST", Group: "/user", Path: "/delete", Auth: "user.delete", Handler: ctrl.delete},
		{Method: "POST", Group: "/user", Path: "/login", Handler: ctrl.login},
		{Method: "GET", Group: "/user", Path: "/current", Handler: ctrl.currentUser},
	}
	for i, hd := range hds {
		middlewares.AuthMap[hd.Group+hd.Path] = &hds[i]
		group.Handle(hd.Method, hd.Path, hd.Handler)
	}
	return ctrl
}
func (ctrl *UserController) byPage(ctx *gin.Context) {
	middlewares.ByPage(ctx, func(pageNumber int, pageSize int) (int64, any, error) {
		return ctrl.userService.ListByPage(pageNumber, pageSize, func(sm *mak.SQLSM) {
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
		*user.InsertedAt = time.Now()
	}
	*user.UpdatedAt = time.Now()
	var rs = ctrl.userService.SaveUpdate(&user)
	affected, err := rs.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	if affected > 0 {
		msg.Ok(ctx, &user)
		return
	}
	msg.Error(ctx, "save failed")
}
func (ctrl *UserController) delete(ctx *gin.Context) {
	var user entity.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	var rs = ctrl.userService.Delete(&user)
	affected, err := rs.RowsAffected()
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
	user, roles, permissions := ctrl.userService.FindByUsernameAndPassword(userInfo["username"], userInfo["password"])
	if user == nil {
		msg.Error(ctx, "username or password error")
		return
	}

	CID, _ := ctx.Cookie("CID")
	if CID != "" {
		utils.DelFromRedis(CID)
	}
	CID = utils.GenerateRandomID()
	maxAge := 60 * 60 * 24
	ctx.SetCookie("CID", CID, maxAge, "/", "", false, true)
	infoMap := map[string]interface{}{
		"user":        user,
		"roles":       roles,
		"permissions": permissions,
	}
	err := utils.SaveObjDataToRedis(CID, infoMap, time.Hour*24)
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	msg.Ok(ctx, user)
}

func (ctrl *UserController) currentUser(ctx *gin.Context) {
	CID, err := ctx.Cookie("CID")
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "Unauthorized",
		})
		return
	}
	infoMap := make(map[string]interface{})
	utils.GetObjDataFromRedis(CID, &infoMap)
	if infoMap == nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "Unauthorized",
		})
		return
	}
	msg.Ok(ctx, infoMap)
}
