package services

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strings"
    "chat2sr/api/models"
    "chat2sr/config"
)


func GenerateSQL(userInput string) (string, error) {


    tableNames := extractTableNames(userInput)
    
    // 如果没有从输入中提取到表名，则使用相关性分析获取相关表
    if len(tableNames) == 0 {
        allTables, err := GetAllTablesWithComments()
        if err != nil {
            return "", fmt.Errorf("获取表失败: %v", err)
        }

        filteredTables := FilterTablesByKeywords(allTables, userInput)

        if len(filteredTables) == 0 {
            return "", fmt.Errorf("无法确定相关表")
        }
        
        for _, tableInfo := range filteredTables {
            tableNames = append(tableNames, tableInfo.Name)
        }
    }


    tableSchemas := make(map[string][]string)
    for _, table := range tableNames {
        columns, err := GetTableSchema(table)
        if err != nil {
            return "", err
        }
        tableSchemas[table] = columns
    }

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
    2. 生成的SQL必须与表结构完全匹配
    3. 只返回SQL语句本身，不要包含任何解释或说明
    4. 不要使用markdown格式
    5. 生成的 SQL 必须与 StarRocks 的语法完全匹配

    用户查询需求：%s`, schemaDesc.String(), userInput)

    requestBody := map[string]interface{}{
        "model":    "deepseek-chat",
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

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+config.AppConfig.DeepSeekAPIKey)

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
        return "", fmt.Errorf("API response contains no choices")
    }

    sql := strings.TrimSpace(response.Choices[0].Message.Content)
    sql = strings.TrimSpace(sql)
    sql = strings.TrimSpace(sql)
    
    // 移除可能的markdown代码块标记
    sql = strings.TrimPrefix(sql, "```sql")
    sql = strings.TrimPrefix(sql, "```")
    sql = strings.TrimSuffix(sql, "```")
    
    // 再次去除空白
    sql = strings.TrimSpace(sql)
    
    return sql, nil
}


func extractTableNames(input string) []string {
    // 查找"需要使用的表:"后面的内容
    tableSection := ""
    if idx := strings.Index(input, "需要使用的表:"); idx != -1 {
        tableSection = input[idx+len("需要使用的表:"):]
        // 如果有下一行，只取到下一行
        if newLineIdx := strings.Index(tableSection, "\n"); newLineIdx != -1 {
            tableSection = tableSection[:newLineIdx]
        }
    } else if idx := strings.Index(input, "需要使用的表："); idx != -1 {
        tableSection = input[idx+len("需要使用的表："):]
        // 如果有下一行，只取到下一行
        if newLineIdx := strings.Index(tableSection, "\n"); newLineIdx != -1 {
            tableSection = tableSection[:newLineIdx]
        }
    }
    
    if tableSection == "" {
        return nil
    }
    
    // 分割表名
    var tableNames []string
    for _, name := range strings.Split(tableSection, ",") {
        name = strings.TrimSpace(name)
        if name != "" {
            tableNames = append(tableNames, name)
        }
    }

    fmt.Println("tableNames :",tableNames)
    
    return tableNames
}

func ProcessQuery(query string) (string, error) {
	requestBody := map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": query,
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

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.AppConfig.DeepSeekAPIKey)

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

	var response models.DeepSeekResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode API response: %v", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("API response contains no choices")
	}

	return response.Choices[0].Message.Content, nil
}