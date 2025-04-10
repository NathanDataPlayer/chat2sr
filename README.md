# 数据智能助手chat2sr

一款通过AI自动生成SQL并执行数据分析的工具，帮助你告别繁杂的取数需求。
结果支持下载、可视化和基本的数据分析。

## 背景介绍

数据开发同学经常面临大量临时取数需求，导致交付效率受限。只要数仓表建设完善，我们可以借助AI能力自动生成SQL逻辑并交付结果，实现数据需求的自助智能化。

## 技术架构

- 前端：H5
- 后端：Go 
- 数据库：Starrocks
- AI引擎：DeepSeek

## 功能演示

1. 用户输入数据需求，如"查询2月份订单的平均值"
   <img width="976" alt="image" src="https://github.com/user-attachments/assets/bf29e8ce-b137-4ce7-9828-178296a44c38" />

2. 系统自动生成对应的SQL代码
   <img width="988" alt="image" src="https://github.com/user-attachments/assets/58298683-166d-425e-a502-164dd6d6da30" />

3. 点击执行SQL后可展示结果，如有需要可以下载结果
   <img width="968" alt="image" src="https://github.com/user-attachments/assets/4b7e8f4d-853c-4171-8237-75c1151874db" />

4. 新增分析报告功能，可以对结果进行分析并给出结论
   <img width="936" alt="image" src="https://github.com/user-attachments/assets/b02c4a8c-9f30-41dd-87b0-9eaa2b1c9793" />


## 快速开始

1. 克隆项目到本地

2. 项目目录下配置 .env 文件，修改相关配置信息
```
DEEPSEEK_API_KEY=<your_ds_key>
DB_HOST=<database_host>
DB_PORT=<database_port>
DB_USER=<database_user>
DB_PASSWORD=<database_password>
DB_NAME=<database_name>
SERVER_PORT=<service_port>
```

3. 启动服务:
```bash
nohup go run main.go &
```

4. 访问服务:
```
http://yourhost:port
```

现在你可以开始使用这个智能数据助手，输入自然语言描述即可自动生成并执行SQL查询。


有任何问题或者想交流的可以加下面微信：
![image](https://github.com/user-attachments/assets/214d289c-549d-4f91-9b26-e0f2f2a32166)


