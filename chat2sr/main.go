package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"errors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// 配置信息
type Config struct {
	DeepSeekAPIKey string
	DBHost         string
	DBPort         string
	DBUser         string
	DBPassword     string
	DBName         string
	ServerPort     string
}

// 请求结构
type QueryRequest struct {
	UserInput string `json:"user_input"`
}

// 执行SQL请求结构
type ExecuteRequest struct {
	SQL string `json:"sql"`
}

// DeepSeek API 请求结构
type DeepSeekRequest struct {
	Messages        []Message      `json:"messages"`
	Model           string         `json:"model"`
	FrequencyPenalty float64       `json:"frequency_penalty"`
	MaxTokens       int            `json:"max_tokens"`
	PresencePenalty float64        `json:"presence_penalty"`
	ResponseFormat  ResponseFormat `json:"response_format"`
	Stop            interface{}    `json:"stop"`
	Stream          bool           `json:"stream"`
	StreamOptions   interface{}    `json:"stream_options"`
	Temperature     float64        `json:"temperature"`
	TopP            float64        `json:"top_p"`
	Tools           interface{}    `json:"tools"`
	ToolChoice      string         `json:"tool_choice"`
	Logprobs        bool           `json:"logprobs"`
	TopLogprobs     interface{}    `json:"top_logprobs"`
}

type Message struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

type ResponseFormat struct {
	Type string `json:"type"`
}

// DeepSeek API 响应结构
type DeepSeekResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

var config Config

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Error loading .env file")
	}

	config = Config{
		DeepSeekAPIKey: os.Getenv("DEEPSEEK_API_KEY"),
		DBHost:         getEnvWithDefault("DB_HOST", "localhost"),
		DBPort:         getEnvWithDefault("DB_PORT", "9030"),
		DBUser:         getEnvWithDefault("DB_USER", "root"),
		DBPassword:     getEnvWithDefault("DB_PASSWORD", ""),
		DBName:         getEnvWithDefault("DB_NAME", "default"),
		ServerPort:     getEnvWithDefault("SERVER_PORT", "8080"),
	}

	if config.DeepSeekAPIKey == "" {
		log.Fatal("DEEPSEEK_API_KEY environment variable is not set")
	}
}

func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(os.Stdout)
	
	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	router.Use(CORSMiddleware())
	router.StaticFile("/", "./index.html")

	api := router.Group("/api")
	{
		api.GET("/health", handleHealth)
		api.POST("/query", handleQuery)
		api.POST("/execute", handleExecute)  
	}

	serverAddr := ":" + config.ServerPort
	fmt.Printf("Server started on port %s\n", config.ServerPort)
	log.Fatal(router.Run(serverAddr))
}

func handleHealth(c *gin.Context) {
	c.String(http.StatusOK, "Service is healthy")
}

func handleQuery(c *gin.Context) {
	var req QueryRequest
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

	sqlQuery, err := generateSQL(req.UserInput)
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

// 新添加的执行SQL处理函数
func handleExecute(c *gin.Context) {
	var req ExecuteRequest
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

	results, err := executeSQL(req.SQL)
	if err != nil {
		log.Printf("Error executing SQL: %s", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute SQL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
	})
}

// getTableSchema获取表结构
func getTableSchema(tableName string) ([]string, error) {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
        config.DBUser,
        config.DBPassword,
        config.DBHost,
        config.DBPort,
        config.DBName)

    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %v", err)
    }
    defer db.Close()

    query := fmt.Sprintf("SHOW COLUMNS FROM %s", tableName)
    rows, err := db.Query(query)
    if err != nil {
        return nil, fmt.Errorf("failed to get schema for table %s: %v", tableName, err)
    }
    defer rows.Close()

    var columns []string
    for rows.Next() {
        var field, fieldType, null, key, extra string
        var defaultValue sql.NullString
        err := rows.Scan(&field, &fieldType, &null, &key, &defaultValue, &extra)
        if err != nil {
            return nil, err
        }
        columns = append(columns, field)
    }

	log.Printf("Retrieved schema for table %s: %v", tableName, columns)
    return columns, nil
}


// 添加新的函数来获取所有表名
func getAllTables() ([]string, error) {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
        config.DBUser,
        config.DBPassword,
        config.DBHost,
        config.DBPort,
        config.DBName)

    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %v", err)
    }
    defer db.Close()

    rows, err := db.Query("SHOW TABLES")
    if err != nil {
        return nil, fmt.Errorf("failed to get tables: %v", err)
    }
    defer rows.Close()

    var tables []string
    for rows.Next() {
        var table string
        if err := rows.Scan(&table); err != nil {
            return nil, err
        }
        tables = append(tables, table)
    }

    return tables, nil
}

// 添加函数来分析表的相关性
func analyzeTableRelevance(userInput string, tables []string) ([]string, error) {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
        config.DBUser,
        config.DBPassword,
        config.DBHost,
        config.DBPort,
        config.DBName)

    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %v", err)
    }
    defer db.Close()

    var relevantTables []string
    userInputLower := strings.ToLower(userInput)

    for _, table := range tables {
        // 获取表的注释
        var tableComment string
        err := db.QueryRow("SELECT table_comment FROM information_schema.tables WHERE table_schema = ? AND table_name = ?",
            config.DBName, table).Scan(&tableComment)
        if err != nil && err != sql.ErrNoRows {
            return nil, err
        }

        // 获取列信息
        columns, err := getTableSchema(table)
        if err != nil {
            return nil, err
        }

        // 检查表名、注释或列名是否与用户输入相关
        if strings.Contains(strings.ToLower(table), userInputLower) ||
           strings.Contains(strings.ToLower(tableComment), userInputLower) {
            relevantTables = append(relevantTables, table)
            continue
        }

        // 检查列名是否相关
        for _, col := range columns {
            if strings.Contains(strings.ToLower(col), userInputLower) {
                relevantTables = append(relevantTables, table)
                break
            }
        }
    }

    // 如果没有找到相关表，返回所有表
    if len(relevantTables) == 0 {
        return tables, nil
    }

    return relevantTables, nil
}


func generateSQL(userInput string) (string, error) {

    
	allTables, err := getAllTables()
    if err != nil {
        return "", fmt.Errorf("failed to get tables: %v", err)
    }

    // 分析相关表
    relevantTables, err := analyzeTableRelevance(userInput, allTables)
    if err != nil {
        return "", fmt.Errorf("failed to analyze table relevance: %v", err)
    }
	
	tableSchemas := make(map[string][]string)
    
    for _, table := range relevantTables {
        columns, err := getTableSchema(table)
        if err != nil {
            return "", err
        }
        tableSchemas[table] = columns
    }

    // 构建表结构说明
    var schemaDesc strings.Builder
    for table, columns := range tableSchemas {
        schemaDesc.WriteString(fmt.Sprintf("%s表字段：\n", table))
        for _, col := range columns {
            schemaDesc.WriteString(fmt.Sprintf("- %s\n", col))
        }
        schemaDesc.WriteString("\n")
    }

	systemPrompt := fmt.Sprintf(`你是一个SQL专家。请严格按照以下数据库表结构生成SQL查询：%s
	要求：
1. 只能使用上述表中实际存在的字段
2. 如果需要的字段不存在，请返回错误提示
3. 生成的SQL必须与表结构完全匹配
4. 只返回SQL语句本身，不要包含任何解释或说明
5. 不要使用markdown格式

用户查询需求：%s`, schemaDesc.String(), userInput)


	requestBody := map[string]interface{}{
		"model":    "deepseek-coder",
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": systemPrompt,
			},
			{
				"role":    "user",
				"content": userInput,
			},
		},
		"temperature": 0.1,
	}

	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.deepseek.com/v1/chat/completions", bytes.NewBuffer(requestJSON))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %v", err)
	}

	log.Printf("Sending request to DeepSeek API: %s", string(requestJSON))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("DEEPSEEK_API_KEY"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode API response: %v", err)
	}

	if len(response.Choices) == 0 {
		return "", errors.New("API response contains no choices")
	}

	log.Printf("Received response from DeepSeek API: %+v", response)

	sqlWithMarkdown := strings.TrimSpace(response.Choices[0].Message.Content)
	sql := sqlWithMarkdown
	sql = strings.TrimPrefix(sql, "```sql")
	sql = strings.TrimPrefix(sql, "```SQL")
	sql = strings.TrimPrefix(sql, "```")
	sql = strings.TrimSuffix(sql, "```")
	sql = strings.TrimSpace(sql)
	
	return sql, nil
}

func executeSQL(sqlQuery string) ([]map[string]interface{}, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", 
		config.DBUser, 
		config.DBPassword, 
		config.DBHost, 
		config.DBPort, 
		config.DBName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	rows, err := db.Query(sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %v", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %v", err)
	}

	var results []map[string]interface{}
	
	count := len(columns)
	values := make([]interface{}, count)
	valuePtrs := make([]interface{}, count)
	
	for rows.Next() {
		for i := range columns {
			valuePtrs[i] = &values[i]
		}
		
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		
		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			
			b, ok := val.([]byte)
			if ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		
		results = append(results, row)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %v", err)
	}
	
	return results, nil
}