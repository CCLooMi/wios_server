package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"wios_server/conf"
	"wios_server/handlers"
	"wios_server/js"
	"wios_server/middlewares"
	"wios_server/task"
	"wios_server/utils"
)

func main() {
	fxa := fx.New(
		conf.Module,
		utils.Module,
		middlewares.Module,
		handlers.Module,
		js.Module,
		mainModule,
		task.Module,
	)
	fxa.Run()
	waitForSignal(fxa)
}

func newGinApp() (*gin.Engine, error) {
	return gin.Default(), nil
}

func startServer(lc fx.Lifecycle, app *gin.Engine, config *conf.Config) {
	if config.DisableServer {
		return
	}
	server := &http.Server{
		Addr:    ":" + config.Port,
		Handler: app,
	}
	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				go func() {
					var err error
					if config.EnableHttps {
						err = server.ListenAndServeTLS(config.CertFile, config.KeyFile)
					} else {
						err = server.ListenAndServe()
					}
					if err != nil {
						log.Println("Error starting server:", err)
						return
					}
				}()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				go func() {
					log.Println("Stopping server...")
					if err := server.Shutdown(ctx); err != nil {
						log.Println("Error stopping server:", err)
						return
					}
				}()
				return nil
			},
		},
	)
}
func waitForSignal(app *fx.App) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.Stop(stopCtx); err != nil {
		log.Println("Failed to stop application:", err)
	}
}

var mainModule = fx.Options(
	fx.Provide(newGinApp),
	fx.Invoke(startServer),
)
