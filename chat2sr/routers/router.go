package routers

import (
    "github.com/gin-gonic/gin"
    "chat2sr/api/handlers"
    "chat2sr/utils"
)

func SetupRouter() *gin.Engine {
    router := gin.Default()
    router.Use(utils.CORSMiddleware())
    router.StaticFile("/", "./index.html")

    api := router.Group("/api")
    {
        api.GET("/health", handlers.HandleHealth)
        api.POST("/query", handlers.HandleNLQuery)   
        api.POST("/execute", handlers.HandleExecute)
    }

    return router
}