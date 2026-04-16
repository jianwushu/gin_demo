// Package main gin-demo service.
//
// @title gin-demo API
// @version 1.0
// @description gin-demo 示例接口文档
// @BasePath /
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"gin-demo/internal/bootstrap"
)

func main() {
	configPath := flag.String("config", "configs/config.yaml", "path to config file")
	flag.Parse()

	if _, err := os.Stat(*configPath); err != nil {
		if os.IsNotExist(err) {
			*configPath = "configs/config.sample.yaml"
		} else {
			panic(fmt.Errorf("stat config file: %w", err))
		}
	}

	app, err := bootstrap.NewApp(*configPath)
	if err != nil {
		panic(err)
	}

	app.Logger.Printf("server starting on %s", app.Config.Server.Address())

	if app.Config.Server.SSL.Enabled {
		if err := app.Server.ListenAndServeTLS(app.Config.Server.SSL.CertFile, app.Config.Server.SSL.KeyFile); err != nil && err != http.ErrServerClosed {
			app.Logger.Fatalf("https server failed: %v", err)
		}
		return
	}

	if err := app.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		app.Logger.Fatalf("http server failed: %v", err)
	}
}
