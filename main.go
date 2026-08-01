package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/lnb/HRPAuth-Backend-Go/config"
	"github.com/lnb/HRPAuth-Backend-Go/controllers"
)

func main() {
	config.Load()

	r := gin.Default()

	exampleCtrl := controllers.NewExampleController()

	r.GET("/example", exampleCtrl.Hello)

	log.Printf("server listening on %s", config.AppConfig.Server.Port)

	r.Run(config.AppConfig.Server.Port)
}
