package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"wios_server/conf"
	"wios_server/handlers/msg"
)

func CorsRevertServer(app *gin.Engine, config *conf.Config) {
	group := app.Group("/proxy")
	proxyMap := make(map[string]*httputil.ReverseProxy)
	var proxyMapMutex sync.Mutex
	if !config.EnableCORS {
		group.Use(func(c *gin.Context) {
			hd := c.Writer.Header()
			for key, value := range config.Header {
				c.Writer.Header().Set(key, value)
			}
			hd.Set("Access-Control-Allow-Methods", c.Request.Method)
			hd.Set("Access-Control-Allow-Origin", c.Request.Header.Get("Origin"))
			hd.Set("Access-Control-Allow-Credentials", "true")
			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(200)
				return
			}
			c.Next()
		})
	}
	group.Any("/*path", func(c *gin.Context) {
		reqPath := c.Param("path")
		corsUrl := "https:/" + reqPath
		targetURL, err := url.Parse(corsUrl)
		if err != nil {
			msg.Error(c, err.Error())
			return
		}
		if v, exists := conf.CorsHostsMap[targetURL.Host]; !exists || !v {
			msg.Error(c, fmt.Sprintf("cors host %s not allow", targetURL.Host))
			return
		}
		hostConfHds := config.HostConf[targetURL.Host].Header
		for key, value := range hostConfHds {
			c.Request.Header.Set(key, value)
		}
		proxyURL, _ := url.Parse(targetURL.Scheme + "://" + targetURL.Host)
		proxyURL.Path = "/"

		proxyMapMutex.Lock()
		proxy, ok := proxyMap[targetURL.String()]
		if !ok {
			proxy = httputil.NewSingleHostReverseProxy(proxyURL)
			proxy.ModifyResponse = func(r *http.Response) error {
				for key, _ := range config.Header {
					r.Header.Del(key)
				}
				r.Header.Del("Access-Control-Allow-Methods")
				r.Header.Del("Access-Control-Allow-Origin")
				r.Header.Del("Access-Control-Allow-Credentials")
				return nil
			}
			proxyMap[targetURL.Host] = proxy
		}
		proxyMapMutex.Unlock()
		c.Request.Host = targetURL.Host
		c.Request.URL.Scheme = targetURL.Scheme
		c.Request.URL.Path = targetURL.Path
		c.Request.URL.Host = targetURL.Host
		c.Request.RequestURI = targetURL.Path + "?" + c.Request.URL.RawQuery
		c.Request.Header.Set("X-Forwarded-Host", c.Request.Header.Get("Host"))
		c.Request.Header.Set("Referer", corsUrl)
		proxy.ServeHTTP(c.Writer, c.Request)
	})
}
