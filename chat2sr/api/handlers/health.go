package handlers

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

func HandleHealth(c *gin.Context) {
    c.String(http.StatusOK, "Service is healthy")
}