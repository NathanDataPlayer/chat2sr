package models

type Config struct {
    DeepSeekAPIKey string
    DBHost         string 
    DBPort         string
    DBUser         string
    DBPassword     string
    DBName         string
    ServerPort     string
}

type QueryRequest struct {
    UserInput string `json:"user_input"`
}

type ExecuteRequest struct {
    SQL string `json:"sql"`
}

type DeepSeekRequest struct {
    Messages         []Message      `json:"messages"`
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

type TableScore struct {
    TableName string
    Score     float64
}


type AnythingLLMChatRequest struct {
	Message     string   `json:"message"`
	ChatHistory []string `json:"chatHistory"`
	WorkspaceId string   `json:"workspaceId"`
}

type AnythingLLMChatResponse struct {
	Response string `json:"response"`
}

type AnythingLLMDocumentRequest struct {
	Name     string `json:"name"`
	Content  string `json:"content"`
	Metadata struct {
		Source string `json:"source"`
	} `json:"metadata"`
}