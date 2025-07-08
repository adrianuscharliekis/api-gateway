package routes

import (
	"api-gateway/handlers"
	"api-gateway/middleware"
	"api-gateway/repository"
	"api-gateway/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(router *gin.Engine, db *gorm.DB) {
	tracelogRepo := repository.NewTracelogRepository(db)
	tracelogService := services.NewTracelogServices(tracelogRepo)
	authHandler := handlers.NewAuthHanlder(tracelogService)
	proxyHandler := handlers.NewProxyHandler(tracelogService)
	router.POST("/auth/login", authHandler.Login)
	router.POST("/generateJWT", handlers.GenerateSignatureHandler)
	secure := router.Group("/secure")
	secure.Use(middleware.JWTAuthMiddleware())
	secure.Use(middleware.BodyCacheMiddleware())
	secure.Any("/*proxyPath", proxyHandler.ProxyHandler)
}
