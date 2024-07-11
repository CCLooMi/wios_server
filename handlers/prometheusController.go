package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"wios_server/middlewares"
)

type PrometheusController struct {
}

func NewPrometheusController(app *gin.Engine) *PrometheusController {
	ctrl := &PrometheusController{}
	group := app.Group("/metrics")
	hds := []middlewares.Auth{
		{Method: "GET", Group: "/metrics", Path: "/*path", Handler: ctrl.metrics},
		{Method: "POST", Group: "/metrics", Path: "/*path", Handler: ctrl.metrics},
		{Method: "PUT", Group: "/metrics", Path: "/*path", Handler: ctrl.metrics},
		{Method: "DELETE", Group: "/metrics", Path: "/*path", Handler: ctrl.metrics},
		{Method: "PATCH", Group: "/metrics", Path: "/*path", Handler: ctrl.metrics},
		{Method: "HEAD", Group: "/metrics", Path: "/*path", Handler: ctrl.metrics},
		{Method: "OPTIONS", Group: "/metrics", Path: "/*path", Handler: ctrl.metrics},
		{Method: "TRACE", Group: "/metrics", Path: "/*path", Handler: ctrl.metrics},
	}
	for i, hd := range hds {
		middlewares.RegisterAuth(&hds[i])
		group.Handle(hd.Method, hd.Path, hd.Handler)
	}
	return ctrl
}
func (ctrl *PrometheusController) metrics(ctx *gin.Context) {
	promhttp.Handler().ServeHTTP(ctx.Writer, ctx.Request)
}
