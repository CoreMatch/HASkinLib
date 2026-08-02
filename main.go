package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/lnb/HRPAuth-Backend-Go/config"
	"github.com/lnb/HRPAuth-Backend-Go/controllers"
	"github.com/lnb/HRPAuth-Backend-Go/database"
)

func main() {
	startupCtrl := controllers.NewStartupController()
	if err := startupCtrl.InitializeConfig(); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	config.Load()
	database.Init()

	if err := startupCtrl.EnsureMigrations(); err != nil {
		log.Fatalf("Failed to ensure database migrations: %v", err)
	}

	r := gin.Default()

	exampleCtrl := controllers.NewExampleController()
	uploadCtrl := controllers.NewUploadController()
	listPreviewCtrl := controllers.NewListPreviewController()

	r.GET("/example", exampleCtrl.Hello)
	r.POST("/texture/upload", uploadCtrl.UploadTexture)
	r.GET("/texture/listpreview", listPreviewCtrl.List)

	log.Printf("server listening on %s", config.AppConfig.Server.Port)

	r.Run(config.AppConfig.Server.Port)
}
