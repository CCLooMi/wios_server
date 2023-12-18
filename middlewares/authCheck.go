package middlewares

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"wios_server/utils"
)

type Auth struct {
	Method  string
	Group   string
	Path    string
	Auth    string
	Handler func(c *gin.Context)
}

var AuthMap = make(map[string]*Auth)

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
	var userInfo map[string]string
	err = utils.GetObjDataFromRedis(cid, &userInfo)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	hasPermission := checkPermission(&userInfo, path)
	if !hasPermission {
		// return 403
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"message": "Forbidden",
		})
		return
	}

	// save user info to context
	c.Set("userInfo", userInfo)

	c.Next()
}

// checkPermission
func checkPermission(userInfo *map[string]string, path string) bool {
	return true
}
