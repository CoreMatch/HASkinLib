package main

import (
	"log"
	"net/http"

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
	r.Use(controllers.RequestIDMiddleware())

	// CORS Middleware
	r.Use(func(c *gin.Context) {
		origin := config.AppConfig.Server.CORS
		if origin == "*" {
			reqOrigin := c.Request.Header.Get("Origin")
			if reqOrigin != "" {
				origin = reqOrigin
			}
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")
		c.Writer.Header().Set("Vary", "Origin")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	exampleCtrl := controllers.NewExampleController()
	uploadCtrl := controllers.NewUploadController()
	listPreviewCtrl := controllers.NewListPreviewController()
	previewFileCtrl := controllers.NewPreviewFileController()
	pullTextureCtrl := controllers.NewPullTextureController()
	deleteTextureCtrl := controllers.NewDeleteTextureController()
	profileCtrl := controllers.NewProfileController()

	r.GET("/example", exampleCtrl.Hello)
	// 专用存活探针：不查数据库、不做重活，供容器 healthcheck 使用。
	// 契约：HA-Contract docs/api/openapi/haskinlib-business.yaml `/health`。
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "ok",
			"data":    gin.H{"status": "ok"},
		})
	})
	r.POST("/texture/upload", uploadCtrl.UploadTexture)
	r.GET("/texture/listpreview", listPreviewCtrl.List)
	r.GET("/texture/preview/:preview_file", previewFileCtrl.Get)
	r.GET("/texture/pull/:hash", pullTextureCtrl.Pull)
	r.POST("/texture/delete", deleteTextureCtrl.DeleteTexture)
	r.GET("/profile/textures", profileCtrl.GetMyTextures)

	log.Printf("server listening on %s", config.AppConfig.Server.Port)

	r.Run(config.AppConfig.Server.Port)
}
