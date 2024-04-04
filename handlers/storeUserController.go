package handlers

import (
	"github.com/CCLooMi/sql-mak/mysql/mak"
	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"wios_server/conf"
	"wios_server/entity"
	"wios_server/handlers/msg"
	"wios_server/middlewares"
	"wios_server/service"
	"wios_server/utils"
)

type StoreUserController struct {
	storeUserService *service.StoreUserService
}

func NewStoreUserController(app *gin.Engine) *StoreUserController {
	ctrl := &StoreUserController{storeUserService: service.NewStoreUserService(conf.Db)}
	group := app.Group("/storeUser")
	hds := []middlewares.Auth{
		{Method: "POST", Group: "/storeUser", Path: "/login", Handler: ctrl.login},
		{Method: "POST", Group: "/storeUser", Path: "/captcha", Handler: ctrl.captcha},
		{Method: "POST", Group: "/storeUser", Path: "/verifyCaptcha", Handler: ctrl.verifyCaptcha},
		{Method: "POST", Group: "/storeUser", Path: "/sendVerifyCodeEmail", Handler: ctrl.sendVerifyCodeEmail},
		{Method: "POST", Group: "/storeUser", Path: "/new", Handler: ctrl.newStoreUser},
		{Method: "POST", Group: "/storeUser", Path: "/byPage", Auth: "storeUser.byPage", Handler: ctrl.byPage},
		{Method: "POST", Group: "/storeUser", Path: "/update", Auth: "#", Handler: ctrl.saveUpdate, AuthCheck: middlewares.StoreAuthCheck},
		{Method: "POST", Group: "/storeUser", Path: "/delete", Auth: "#", Handler: ctrl.delete, AuthCheck: middlewares.StoreAuthCheck},
		{Method: "GET", Group: "/storeUser", Path: "/current", Auth: "#", Handler: ctrl.currentStoreUser, AuthCheck: middlewares.StoreAuthCheck},
		{Method: "GET", Group: "/storeUser", Path: "/logout", Auth: "#", Handler: ctrl.logout, AuthCheck: middlewares.StoreAuthCheck},
	}
	for i, hd := range hds {
		middlewares.RegisterAuth(&hds[i])
		group.Handle(hd.Method, hd.Path, hd.Handler)
	}
	return ctrl
}
func (ctrl *StoreUserController) byPage(ctx *gin.Context) {
	middlewares.ByPage(ctx, func(page *middlewares.Page) (int64, any, error) {
		return ctrl.storeUserService.ListByPage(page.PageNumber, page.PageSize, func(sm *mak.SQLSM) {
			sm.SELECT("*").FROM(entity.StoreUser{}, "u")
		})
	})
}

type redisStore struct {
	Expire time.Duration
	base64Captcha.Store
}

func (c *redisStore) Set(id string, value string) error {
	return utils.SaveKVToRedis(id, value, c.Expire)
}
func (c *redisStore) Get(key string, clear bool) string {
	v, _ := utils.GetValueFromRedis(key)
	if clear {
		utils.DelFromRedis(key)
	}
	return v
}
func (c *redisStore) Verify(id, answer string, clear bool) bool {
	v := c.Get(id, clear)
	if v == "" {
		return false
	}
	return v == answer
}

type captchaConfig struct {
	CaptchaType   string                       `json:"captchaType"`
	DriverAudio   *base64Captcha.DriverAudio   `json:"audio"`
	DriverString  *base64Captcha.DriverString  `json:"string"`
	DriverChinese *base64Captcha.DriverChinese `json:"chinese"`
	DriverMath    *base64Captcha.DriverMath    `json:"math"`
	DriverDigit   *base64Captcha.DriverDigit   `json:"digit"`
}

var captchaStore = &redisStore{
	Expire: 600 * time.Second,
}

func (ctrl *StoreUserController) captcha(ctx *gin.Context) {
	var conf captchaConfig
	if err := ctx.ShouldBindJSON(&conf); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	var driver base64Captcha.Driver
	switch conf.CaptchaType {
	case "audio":
		driver = conf.DriverAudio
	case "string":
		driver = conf.DriverString.ConvertFonts()
	case "math":
		driver = conf.DriverMath.ConvertFonts()
	case "chinese":
		driver = conf.DriverChinese.ConvertFonts()
	default:
		driver = conf.DriverDigit
	}
	captchaInstance := base64Captcha.NewCaptcha(driver, captchaStore)
	id, captchaB64, _, err := captchaInstance.Generate()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	msg.Ok(ctx, gin.H{
		"id":      id,
		"captcha": captchaB64,
	})
}
func (ctrl *StoreUserController) verifyCaptcha(ctx *gin.Context) {
	var m map[string]string
	if err := ctx.ShouldBindJSON(&m); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	ok := captchaStore.Verify(m["id"], m["code"], false)
	if ok {
		msg.Ok(ctx, nil)
		return
	}
	msg.Error(ctx, "invalid answer")
}
func (ctrl *StoreUserController) sendVerifyCodeEmail(ctx *gin.Context) {
	var info map[string]string
	if err := ctx.ShouldBindJSON(&info); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	captchaId := info["captchaId"]
	code := info["code"]
	if !captchaStore.Verify(captchaId, code, true) {
		msg.Error(ctx, "invalid answer")
		return
	}
	tmp := conf.SysCfg["code.verify.template"].(string)
	code = utils.GenRandomNum(6)
	info["code"] = code
	info["year"] = strconv.Itoa(time.Now().Year())
	body, err := utils.ApplyTemplate(&tmp, "code.verify.template", info)
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	email := info["email"]
	err = utils.SendEmail("Verification Code From WiOS Group", &body, email)
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	err = captchaStore.Set(email, code)
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	msg.Ok(ctx, "ok")
}
func (ctrl *StoreUserController) newStoreUser(ctx *gin.Context) {
	var m map[string]string
	if err := ctx.ShouldBindJSON(&m); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	email := m["email"]
	code := m["code"]
	if !captchaStore.Verify(email, code, true) {
		msg.Error(ctx, "invalid verify code!")
		return
	}
	password := m["password"]
	avatar := m["avatar"]
	storeUser := entity.StoreUser{
		Email:  &email,
		Avatar: &avatar,
		Seed:   utils.RandomBytes(8),
	}
	if ctrl.storeUserService.CheckExist(&storeUser) {
		msg.Error(ctx, "user exists")
		return
	}
	storeUser.Password = utils.SHA_256(password, storeUser.Seed)
	r := ctrl.storeUserService.SaveOrUpdate(&storeUser)
	rows, err := r.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	if rows == 0 {
		msg.Error(ctx, "save failed")
		return
	}
	msg.Ok(ctx, &storeUser)
}
func (ctrl *StoreUserController) saveUpdate(ctx *gin.Context) {
	var storeUser entity.StoreUser
	if err := ctx.ShouldBindJSON(&storeUser); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	if storeUser.Id == nil {
		if ctrl.storeUserService.CheckExist(&storeUser) {
			msg.Error(ctx, "username email or phone exists")
			return
		}
	} else {
		currentUserInfo := ctx.MustGet(middlewares.StoreUserInfoKey).(*middlewares.StoreUserInfo)
		currentUser := currentUserInfo.User
		if *currentUser.Id != *storeUser.Id {
			msg.Error(ctx, "permission denied!")
			return
		}
	}
	if storeUser.Seed == nil {
		storeUser.Seed = utils.RandomBytes(8)
		storeUser.Password = utils.SHA_256(storeUser.Password, storeUser.Seed)
	}
	var rs = ctrl.storeUserService.SaveOrUpdate(&storeUser)
	_, err := rs.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	msg.Ok(ctx, &storeUser)
}
func (ctrl *StoreUserController) delete(ctx *gin.Context) {
	var storeUser entity.StoreUser
	if err := ctx.ShouldBindJSON(&storeUser); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	var rs = ctrl.storeUserService.Delete(&storeUser)
	affected, err := rs.RowsAffected()
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	if affected > 0 {
		msg.Ok(ctx, &storeUser)
		return
	}
	msg.Error(ctx, "delete failed")
}
func (ctrl *StoreUserController) login(ctx *gin.Context) {
	var userInfo map[string]string
	if err := ctx.ShouldBindJSON(&userInfo); err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	storeUser := ctrl.storeUserService.FindByUsernameAndPassword(userInfo["username"], userInfo["password"])
	if storeUser.Id == nil {
		msg.Error(ctx, "username password error")
		return
	}
	SID, _ := ctx.Cookie(middlewares.StoreSessionIDKey)
	if SID != "" {
		utils.DelFromRedis(SID)
	}
	SID = utils.GenerateRandomID()
	domain := utils.RemoveDomainPort(ctx.Request.Host)
	maxAge := 60 * 60 * 24
	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     middlewares.StoreSessionIDKey,
		Value:    url.QueryEscape(SID),
		MaxAge:   maxAge,
		Path:     "/",
		Domain:   domain,
		SameSite: http.SameSiteNoneMode,
		Secure:   true,
		HttpOnly: true,
	})
	infoMap := map[string]interface{}{
		"user": storeUser,
	}
	err := utils.SaveObjDataToRedis(SID, infoMap, time.Hour*24)
	if err != nil {
		msg.Error(ctx, err.Error())
		return
	}
	msg.Ok(ctx, &storeUser)
}
func (ctrl *StoreUserController) currentStoreUser(ctx *gin.Context) {
	userInfo := ctx.MustGet(middlewares.StoreUserInfoKey).(*middlewares.StoreUserInfo)
	msg.Ok(ctx, userInfo)
}
func (ctrl *StoreUserController) logout(ctx *gin.Context) {
	SID, err := ctx.Cookie(middlewares.StoreSessionIDKey)
	if err != nil || SID == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"msg": "unauthorized"})
		return
	}
	utils.DelFromRedis(SID)
	msg.Ok(ctx, nil)
}
