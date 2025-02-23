package handlers

import (
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/fx"
	"wios_server/conf"
	"wios_server/middlewares"
	"wios_server/utils"
)

var Module = fx.Options(
	fx.Invoke(
		use,
		HandleWebSocket,
		HandleFileUpload,
		ServerUploadFile,
		NewPrometheusController,
		NewUserController,
		NewMenuController,
		NewOrgController,
		NewRoleController,
		NewUploadController,
		NewFilesController,
		NewApiController,
		NewConfigController,
		NewStoreUserController,
		NewWppController,
		NewWppEventController,
		NewWppStoryController,
		NewAiAssistantController,
		CorsRevertServer,
		ServerStaticDir,
	),
)

func use(app *gin.Engine, ut *utils.Utils, config *conf.Config, ac *middlewares.AuthChecker) {
	app.Use(func(c *gin.Context) {
		c.Set("json", jsoniter.ConfigFastest)
		c.Next()
	}, func(c *gin.Context) {
		middlewares.ApplyConfig(c, config)
	}, ac.AuthCheck)
}
