package middlewares

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func AuthCheck(c *gin.Context) {
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
	userInfo := getUserInfoFromRedis(cid)
	if userInfo == nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	hasPermission := checkPermission(userInfo, c.Request.URL.Path)
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
func getUserInfoFromRedis(cid string) (userInfo *map[string]interface{}) {
	// get user info from redis
	return nil
}

// checkPermission
func checkPermission(userInfo *map[string]interface{}, path string) bool {
	return true
}
