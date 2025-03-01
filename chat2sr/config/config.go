package config

import (
    "os"
    "log"
    "github.com/joho/godotenv"
)

// 配置信息结构
type Config struct {
    DeepSeekAPIKey string
    DBHost         string
    DBPort         string
    DBUser         string
    DBPassword     string
    DBName         string
    ServerPort     string
}

var AppConfig Config

// 初始化配置
func Init() {
    // 加载 .env 文件
    if err := godotenv.Load(); err != nil {
        log.Printf("Error loading .env file: %v", err)
    }

    AppConfig = Config{
        DeepSeekAPIKey: os.Getenv("DEEPSEEK_API_KEY"),
        DBHost:         GetEnvWithDefault("DB_HOST", ""),
        DBPort:         GetEnvWithDefault("DB_PORT", "9030"),
        DBUser:         GetEnvWithDefault("DB_USER", ""),
        DBPassword:     GetEnvWithDefault("DB_PASSWORD", ""),
        DBName:         GetEnvWithDefault("DB_NAME", "default"),
        ServerPort:     GetEnvWithDefault("SERVER_PORT", "8080"),
    }

    if AppConfig.DeepSeekAPIKey == "" {
        log.Fatal("DEEPSEEK_API_KEY environment variable is not set")
    }
}

// 获取环境变量，带默认值
func GetEnvWithDefault(key, defaultValue string) string {
    value := os.Getenv(key)
    if value == "" {
        return defaultValue
    }
    return value
}