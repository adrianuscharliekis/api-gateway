package routes

import (
	"api-gateway/handlers"
	"api-gateway/middleware"
	"api-gateway/repository"
	"api-gateway/services"
	"api-gateway/utils"
	"database/sql"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, db *sql.DB) {
	tracelogRepo := repository.NewTracelogRepository(db)
	productRepo := repository.NewProductRepository(db)
	tracelogService := services.NewTracelogServices(tracelogRepo)
	externalIDStore := utils.NewExternalIDStore()
	productServices := services.NewProductService(productRepo, tracelogService)
	authHandler := handlers.NewAuthHandler(tracelogService, externalIDStore, productServices)
	proxyHandler := handlers.NewProxyHandler(tracelogService)
	router.POST("/auth/login", authHandler.Login)
	router.POST("/generateJWT", handlers.GenerateSignatureHandler)
	secure := router.Group("/secure")
	secure.Use(middleware.JWTAuthMiddleware())
	secure.Use(middleware.BodyCacheMiddleware())
	secure.Any("/*proxyPath", proxyHandler.ProxyHandler)
}
