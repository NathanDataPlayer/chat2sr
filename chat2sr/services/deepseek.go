package services

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strings"
    "sort"
    "chat2sr/api/models"
    "chat2sr/config"
)

func AnalyzeTableRelevance(userInput string, tables []string) ([]string, error) {
    var tableScores []models.TableScore
    userInputLower := strings.ToLower(userInput)

    for _, table := range tables {
        score := 0.0
        
        columns, err := GetTableSchema(table)
        if err != nil {
            return nil, err
        }

        if strings.ToLower(table) == userInputLower {
            score += 10.0
        } else if strings.Contains(strings.ToLower(table), userInputLower) {
            score += 5.0
        }

        for _, col := range columns {
            if strings.ToLower(col) == userInputLower {
                score += 6.0
            } else if strings.Contains(strings.ToLower(col), userInputLower) {
                score += 3.0
            }
        }

        tableScores = append(tableScores, models.TableScore{TableName: table, Score: score})
    }

    sort.Slice(tableScores, func(i, j int) bool {
        return tableScores[i].Score > tableScores[j].Score
    })

    threshold := 3.0
    var relevantTables []string
    for _, ts := range tableScores {
        if ts.Score >= threshold {
            relevantTables = append(relevantTables, ts.TableName)
        }
    }

    if len(relevantTables) == 0 {
        return tables, nil
    }

    return relevantTables, nil
}

func GenerateSQL(userInput string) (string, error) {
    allTables, err := GetAllTables()
    if err != nil {
        return "", fmt.Errorf("failed to get tables: %v", err)
    }

    relevantTables, err := AnalyzeTableRelevance(userInput, allTables)
    if err != nil {
        return "", fmt.Errorf("failed to analyze table relevance: %v", err)
    }

    tableSchemas := make(map[string][]string)
    for _, table := range relevantTables {
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
    2. 如果需要的字段不存在，请返回错误提示
    3. 生成的SQL必须与表结构完全匹配
    4. 只返回SQL语句本身，不要包含任何解释或说明
    5. 不要使用markdown格式
    6. 生成的 SQL 必须与 StarRocks 的语法完全匹配

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

    sqlWithMarkdown := strings.TrimSpace(response.Choices[0].Message.Content)
    sql := sqlWithMarkdown
    sql = strings.TrimPrefix(sql, "```sql")
    sql = strings.TrimPrefix(sql, "```SQL")
    sql = strings.TrimPrefix(sql, "```")
    sql = strings.TrimSuffix(sql, "```")
    sql = strings.TrimSpace(sql)
    
    return sql, nil
}