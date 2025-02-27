# chat2sr
数据智能助手：通过输入需求自动生成SQL代码，并执行计算和展示结果。让你告别繁杂的取数需求。

对于数据开发同学，经常会遇到临时取数的需求，如果需求一堆积，很容易造成交付不及时。

其实对于这类需求，只要数仓表建设够完善，完全可以借助AI的能力来自动写SQL逻辑并交付结果，做到数据需求的线上智能化。

整个项目的前后端组件：
-前端：H5
-后端：Go
-数据库：Starrocks
-AI引擎：DeepSeek


基础效果演示:

用户端输入数据需求：比如想看 “消费最多的客户”
<img width="941" alt="image" src="https://github.com/user-attachments/assets/7558914e-d16a-43f7-b019-a9e99b03e3bc" />

该服务自动生成SQL代码，点击执行SQL，会展示出结果
<img width="907" alt="image" src="https://github.com/user-attachments/assets/bbf433e7-de51-468a-b358-00ef3d90fbd5" />


项目运行也比较简单：
只要clone项目到本地，然后修改.env配置文件里面的信息

DEEPSEEK_API_KEY=更新为自己的DS KEY
DB_HOST=DB的Host
DB_PORT=DB的端口
DB_USER=DB用户名
DB_PASSWORD=
DB_NAME=DB名字
SERVER_PORT=服务端口

修改完后，直接 nohup go run main.go & 执行即可。

服务起来后，可以在用户端通过http://yourhost:port进行访问
