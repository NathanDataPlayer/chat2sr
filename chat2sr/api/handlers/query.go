package handlers

import (
    "log"
    "net/http"
    "strings"
    "chat2sr/api/models"
    "chat2sr/services"
    "github.com/gin-gonic/gin"
    "fmt"
)

// HandleNLQuery 处理自然语言转SQL的查询
func HandleNLQuery(c *gin.Context) {
    fmt.Println("Start :")
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
    
    // 1. 获取所有表及其注释
    allTables, err := services.GetAllTablesWithComments()
    if err != nil {
        log.Printf("Error getting tables: %s", err.Error())
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get database tables"})
        return
    }
    
    // 2. 使用关键词匹配筛选可能相关的表
    filteredTables := services.FilterTablesByKeywords(allTables, req.UserInput)
    log.Printf("Filtered tables by keywords: %v", filteredTables)
    
    // 3. 获取筛选后表的详细结构信息
    var tablesInfo strings.Builder
    for _, table := range filteredTables {
        columns, err := services.GetTableSchemaWithComments(table.Name)
        if err != nil {
            log.Printf("Error getting schema for table %s: %s", table.Name, err.Error())
            continue
        }
        
        tablesInfo.WriteString(fmt.Sprintf("表名: %s", table.Name))
        if table.Comment != "" {
            tablesInfo.WriteString(fmt.Sprintf(" (说明: %s)", table.Comment))
        }
        tablesInfo.WriteString("\n字段列表:\n")
        
        for _, col := range columns {
            fieldDesc := fmt.Sprintf("- %s (%s)", col["name"], col["type"])
            if col["comment"] != "" {
                fieldDesc += fmt.Sprintf(" 说明: %s", col["comment"])
            }
            tablesInfo.WriteString(fieldDesc + "\n")
        }
        tablesInfo.WriteString("\n")
    }

    fmt.Println("tablesInfo.String() :",tablesInfo.String())
    
    // 4. 调用LLM服务识别需要的表
    log.Printf("Identifying required tables using LLM...")
    llmPrompt := fmt.Sprintf(`分析以下用户需求，从这些表中选择最相关的表。
请结合表的结构、字段含义和表的说明，选择最适合回答用户问题的表。
只返回表名，多个表用逗号分隔。

数据库表结构:
%s

用户需求: %s`, tablesInfo.String(), req.UserInput)

    tablesResponse, err := services.ProcessQuery(llmPrompt)
    if err != nil {
        log.Printf("Error identifying tables with LLM: %s", err.Error())
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to identify required tables"})
        return
    }
    log.Printf("LLM identified tables: %s", tablesResponse)
    
    // 5. 将用户输入和识别的表信息一起传递给SQL生成服务
    enrichedInput := fmt.Sprintf("用户需求: %s\n需要使用的表: %s", req.UserInput, tablesResponse)
    log.Printf("Generating SQL with enriched input: %s", enrichedInput)
    
    sqlQuery, err := services.GenerateSQL(enrichedInput)
    if err != nil {
        log.Printf("Error generating SQL: %s", err.Error())
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate SQL"})
        return
    }
    log.Printf("Generated SQL query: %s", sqlQuery)

    response := gin.H{
        "sql": sqlQuery,
        "tables": strings.Split(strings.ReplaceAll(tablesResponse, " ", ""), ","),
    }
    
    log.Printf("Preparing response data: %+v", response)
    c.JSON(http.StatusOK, response)
    log.Printf("Response sent with status 200 and data: %+v", response)
}