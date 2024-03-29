package middlewares

import (
	"crypto/sha1"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"wios_server/entity"
	"wios_server/utils"
)

type Auth struct {
	id        *string
	Method    string
	Group     string
	Path      string
	Auth      string
	Handler   func(c *gin.Context)
	AuthCheck func(c *gin.Context)
}

func (a *Auth) GetId() string {
	if a.id != nil {
		return *a.id
	}
	hash := sha1.Sum([]byte(a.Method + a.Group + a.Path))
	id := hex.EncodeToString(hash[:])
	a.id = &id
	return *a.id
}

var authMap = make(map[string]*Auth)
var AuthList = make([]*Auth, 0)

type UserInfo struct {
	User        *entity.User    `json:"user"`
	Roles       []entity.Role   `json:"roles"`
	Permissions map[string]bool `json:"permissions"`
}
type StoreUserInfo struct {
	User *entity.User `json:"user"`
}

func RegisterAuths(auths ...*Auth) {
	for _, auth := range auths {
		RegisterAuth(auth)
	}
}
func RegisterAuth(auth *Auth) {
	a := authMap[auth.GetId()]
	if a == nil {
		authMap[auth.Group+auth.Path] = auth
		AuthList = append(AuthList, auth)
	}
}

const UserInfoKey = "userInfo"
const UserSessionIDKey = "CID"

func AuthCheck(c *gin.Context) {
	path := c.Request.URL.Path
	auth := authMap[path]
	if auth == nil || auth.Auth == "" {
		c.Next()
		return
	}
	if auth.AuthCheck != nil {
		auth.AuthCheck(c)
		return
	}
	// check CID value
	cid, err := c.Cookie(UserSessionIDKey)
	if err != nil {
		// return 401
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	// get user info from redis by CID
	var userInfo = UserInfo{}
	err = utils.GetObjDataFromRedis(cid, &userInfo)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "Unauthorized",
		})
		return
	}
	if strings.HasPrefix(auth.Auth, "#") {
		if auth.Auth != "#" && auth.Auth != ("#"+userInfo.User.Username) {
			// return 403
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message": "Forbidden",
			})
			return
		}
	} else if userInfo.User.Username != "root" {
		hasPermission := checkPermission(&userInfo, auth)
		if !hasPermission {
			// return 403
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message": "Forbidden",
			})
			return
		}
	}

	// save user info to context
	c.Set(UserInfoKey, &userInfo)

	c.Next()
}

const StoreUserInfoKey = "storeUserInfo"
const StoreSessionIDKey = "SID"

func StoreAuthCheck(c *gin.Context) {
	path := c.Request.URL.Path
	auth := authMap[path]
	if auth == nil || auth.Auth == "" {
		c.Next()
		return
	}
	// check SID value
	sid, err := c.Cookie(StoreSessionIDKey)
	if err != nil {
		// return 401
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "Unauthorized",
		})
		return
	}
	// get user info from redis by SID
	var storeUserInfo = StoreUserInfo{}
	err = utils.GetObjDataFromRedis(sid, &storeUserInfo)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "Unauthorized",
		})
		return
	}
	// save user info to context
	c.Set(StoreUserInfoKey, &storeUserInfo)
	c.Next()
}

// checkPermission
func checkPermission(userInfo *UserInfo, auth *Auth) bool {
	return userInfo.Permissions[auth.GetId()]
}
