package main

import (
	"api-gateway/config"
	"api-gateway/model"
	"api-gateway/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	config.Startup()
	config.ConnectDB()
	config.DB.AutoMigrate(&model.Tracelog{})
	routes.RegisterRoutes(r, config.DB)
	r.Run(":" + config.Config.Server["port"].(string))
}
