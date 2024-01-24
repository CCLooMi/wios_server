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
	id      *string
	Method  string
	Group   string
	Path    string
	Auth    string
	Handler func(c *gin.Context)
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

var AuthMap = make(map[string]*Auth)

type UserInfo struct {
	User        *entity.User  `json:"user"`
	Roles       []entity.Role `json:"roles"`
	Permissions []string      `json:"permissions"`
}

func AuthCheck(c *gin.Context) {
	path := c.Request.URL.Path
	auth := AuthMap[path]
	if auth == nil || auth.Auth == "" {
		c.Next()
		return
	}
	// check CID value
	cid, err := c.Cookie("CID")
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
		hasPermission := checkPermission(&userInfo, path)
		if !hasPermission {
			// return 403
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"message": "Forbidden",
			})
			return
		}
	}

	// save user info to context
	c.Set("userInfo", &userInfo)

	c.Next()
}

// checkPermission
func checkPermission(userInfo *UserInfo, path string) bool {
	return true
}
