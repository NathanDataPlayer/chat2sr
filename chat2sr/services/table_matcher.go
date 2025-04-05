package services

import (
    "strings"
    "sort"
)

// FilterTablesByKeywords 通过关键词匹配筛选可能相关的表
func FilterTablesByKeywords(tables []TableInfo, userInput string) []TableInfo {
    // 如果表数量不多，直接返回所有表
    if len(tables) <= 10 {
        return tables
    }
    
    // 将用户输入转换为小写
    userInput = strings.ToLower(userInput)
    
    // 计算每个表与用户输入的相关度
    type tableScore struct {
        table TableInfo
        score float64
    }
    
    var scores []tableScore
    
    for _, table := range tables {
        tableLower := strings.ToLower(table.Name)
        commentLower := strings.ToLower(table.Comment)
        
        // 初始分数
        score := 0.0
        
        // 1. 检查表名与用户输入的匹配度
        if strings.Contains(userInput, tableLower) {
            // 用户输入直接包含表名，这是最强的关联
            score += 1.0
        } else if strings.Contains(tableLower, userInput) {
            // 表名包含整个用户输入
            score += 0.8
        }
        
        // 2. 检查表注释与用户输入的匹配度
        if commentLower != "" {
            if strings.Contains(userInput, commentLower) {
                score += 0.9
            } else if strings.Contains(commentLower, userInput) {
                score += 0.7
            }
            
            // 检查注释中的关键词
            userWords := strings.Fields(userInput)
            for _, word := range userWords {
                if len(word) > 2 && strings.Contains(commentLower, word) {
                    score += 0.3
                }
            }
        }
        
        // 3. 将表名拆分为单词（按下划线或驼峰命名）
        tableWords := splitTableName(tableLower)
        
        // 4. 计算用户输入中包含多少个表名单词
        matchCount := 0
        for _, word := range tableWords {
            if len(word) > 1 && strings.Contains(userInput, word) {
                matchCount++
            }
        }
        
        if matchCount > 0 {
            // 计算匹配率
            matchRatio := float64(matchCount) / float64(len(tableWords))
            score += 0.4 * matchRatio
        } else {
            // 5. 检查用户输入的单词是否出现在表名中
            userWords := strings.Fields(userInput)
            for _, word := range userWords {
                if len(word) > 2 && strings.Contains(tableLower, word) {
                    score += 0.2
                    break
                }
            }
        }
        
        // 只有得分大于0的表才加入结果
        if score > 0 {
            scores = append(scores, tableScore{table, score})
        }
    }
    
    // 按得分排序
    sort.Slice(scores, func(i, j int) bool {
        return scores[i].score > scores[j].score
    })
    
    // 如果没有匹配的表，返回所有表中的前10个
    if len(scores) == 0 {
        if len(tables) > 10 {
            return tables[:10]
        }
        return tables
    }
    
    // 返回得分最高的最多10个表
    var result []TableInfo
    for i, ts := range scores {
        if i >= 10 {
            break
        }
        result = append(result, ts.table)
    }
    
    return result
}

// splitTableName 将表名分解为单词
func splitTableName(tableName string) []string {
    // 处理下划线分隔的表名
    if strings.Contains(tableName, "_") {
        parts := strings.Split(tableName, "_")
        var words []string
        for _, part := range parts {
            if part != "" {
                words = append(words, part)
            }
        }
        return words
    }
    
    // 处理驼峰命名的表名
    var words []string
    var currentWord strings.Builder
    
    for i, char := range tableName {
        if i > 0 && (char >= 'A' && char <= 'Z') {
            words = append(words, strings.ToLower(currentWord.String()))
            currentWord.Reset()
        }
        currentWord.WriteRune(char)
    }
    
    if currentWord.Len() > 0 {
        words = append(words, strings.ToLower(currentWord.String()))
    }
    
    return words
}