package main

import (
    "fmt"
    "log"
    "chat2sr/config"
    "chat2sr/logs"
    "chat2sr/routers"
)

func main() {
    // 初始化日志
    logs.Init()

    // 初始化配置
    config.Init()

    // 设置路由
    router := routers.SetupRouter()

    // 启动服务器
    serverAddr := ":" + config.AppConfig.ServerPort
    fmt.Printf("Server started on port %s\n", config.AppConfig.ServerPort)
    log.Fatal(router.Run(serverAddr))
}