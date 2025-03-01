package handlers

import (
    "log"
    "net/http"
    "strings"
    "chat2sr/api/models"
    "chat2sr/services"
    "github.com/gin-gonic/gin"
)

func HandleQuery(c *gin.Context) {
    var req models.QueryRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        log.Printf("Invalid request: %s", err.Error())
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
        return
    }

    if strings.TrimSpace(req.UserInput) == "" {
        log.Printf("Empty user input received")
        c.JSON(http.StatusBadRequest, gin.H{"error": "Please provide a query description"})
        return
    }
    log.Printf("Received user input: %s", req.UserInput)

    sqlQuery, err := services.GenerateSQL(req.UserInput)
    if err != nil {
        log.Printf("Error generating SQL: %s", err.Error())
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate SQL"})
        return
    }
    log.Printf("Generated SQL query: %s", sqlQuery)

    c.JSON(http.StatusOK, gin.H{
        "sql": sqlQuery,
    })
}