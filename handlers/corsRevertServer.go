package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http/httputil"
	"net/url"
	"sync"
	"wios_server/conf"
	"wios_server/handlers/msg"
)

func CorsRevertServer(app *gin.Engine) {
	group := app.Group("/proxy")
	proxyMap := make(map[string]*httputil.ReverseProxy)
	var proxyMapMutex sync.Mutex
	group.Use(func(c *gin.Context) {
		for key, value := range conf.Cfg.Header {
			c.Writer.Header().Set(key, value)
		}
		c.Writer.Header().Set("Access-Control-Allow-Methods", c.Request.Method)
		c.Writer.Header().Set("Access-Control-Allow-Origin", c.Request.Header.Get("Origin"))
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}
		c.Next()
	})
	group.Any("/*path", func(c *gin.Context) {
		reqPath := c.Param("path")
		corsUrl := "https:/" + reqPath
		targetURL, err := url.Parse(corsUrl)
		if err != nil {
			msg.Error(c, err.Error())
			return
		}
		hostConfHds := conf.Cfg.HostConf[targetURL.Host].Header
		for key, value := range hostConfHds {
			c.Request.Header.Set(key, value)
		}
		proxyURL, _ := url.Parse(targetURL.Scheme + "://" + targetURL.Host)
		proxyURL.Path = "/"

		proxyMapMutex.Lock()
		proxy, ok := proxyMap[targetURL.String()]
		if !ok {
			proxy = httputil.NewSingleHostReverseProxy(proxyURL)
			proxyMap[targetURL.Host] = proxy
		}
		proxyMapMutex.Unlock()
		for key, values := range c.Request.Header {
			for _, value := range values {
				c.Request.Header.Set(key, value)
			}
		}
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
