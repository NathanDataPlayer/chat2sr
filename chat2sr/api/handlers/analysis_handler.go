package handlers

import (
	"bytes"
	"chat2sr/config"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	
	"github.com/gin-gonic/gin"
)

// AnalysisRequest 分析请求
type AnalysisRequest struct {
	Query  string        `json:"query"`
	SQL    string        `json:"sql"`
	Result []interface{} `json:"result"`
}

// AnalysisResponse 分析响应
type AnalysisResponse struct {
	Analysis string `json:"analysis"`
	Error    string `json:"error,omitempty"`
}

// HandleAnalysis 处理分析请求
func HandleAnalysis(c *gin.Context) {
	// 解析请求
	var req AnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 调用DeepSeek API生成分析报告
	analysis, err := generateAnalysisReport(req.Query, req.SQL, req.Result)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AnalysisResponse{
			Error: fmt.Sprintf("生成分析报告失败: %v", err),
		})
		return
	}

	// 返回分析报告
	c.JSON(http.StatusOK, AnalysisResponse{
		Analysis: analysis,
	})
}

// generateAnalysisReport 生成分析报告
func generateAnalysisReport(query, sql string, result []interface{}) (string, error) {
	// 构建请求DeepSeek API的数据
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return "", err
	}

	// 构建提示词
	prompt := fmt.Sprintf(`
你是一位数据分析专家，请根据以下信息生成一份专业的数据分析报告：

1. 用户查询: %s
2. 执行的SQL: %s
3. 查询结果: %s

请提供以下内容：
1. 数据概览：简要描述数据的基本情况
2. 关键发现：指出数据中的重要趋势、模式或异常
3. 详细分析：针对用户查询进行深入分析
4. 结论和建议：总结分析结果并提供actionable的建议

请直接使用HTML格式输出，不要使用Markdown。使用适当的HTML标签（如<h1>, <h2>, <ul>, <li>, <strong>, <em>等）来格式化内容，确保报告专业、简洁且有洞察力。
`, query, sql, string(resultJSON))

	// 构建请求体
	requestBody, err := json.Marshal(map[string]interface{}{
		"model": "deepseek-chat",
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.7,
		"max_tokens":  2000,
	})
	if err != nil {
		return "", err
	}

	// 创建请求
	req, err := http.NewRequest("POST", "https://api.deepseek.com/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.AppConfig.DeepSeekAPIKey)

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// 解析响应
	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("API返回空响应")
	}

	// 返回生成的分析报告
	return response.Choices[0].Message.Content, nil
}