package handlers

import (
    "log"
    "net/http"
    "strings"
    "github.com/gin-gonic/gin"
    "chat2sr/services"
    "chat2sr/api/models"
)

func HandleExecute(c *gin.Context) {
    var req models.ExecuteRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        log.Printf("Invalid request: %s", err.Error())
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }

    if strings.TrimSpace(req.SQL) == "" {
        log.Printf("Empty SQL received")
        c.JSON(http.StatusBadRequest, gin.H{"error": "Please provide SQL statement"})
        return
    }

    log.Printf("Executing SQL: %s", req.SQL)

    results, err := services.ExecuteSQL(req.SQL)
    if err != nil {
        log.Printf("Error executing SQL: %s", err.Error())
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute SQL"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "results": results,
    })
}