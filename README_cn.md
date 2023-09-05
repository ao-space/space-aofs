# space-aofs

[English](./README.md) | 简体中文

AOFS 为傲空间服务器端提供傲空间文件服务的内部模块，通过 space-gateway 和 space-nginx 模块提供文件访问接口，包括文件列表、文件上传、文件下载及文件管理等接口。

## 编译构建

### 环境准备

准备好 docker环境.

### 源码下载

请按下面步骤下载源码:

- 创建本地模块目录，执行命令: `mkdir space-aofs`
- 进入模块目录: `cd ./space-aofs`
- 下载源码: `git clone git@github.com:ao-space/space-aofs.git .`

### 容器镜像构建

进入模块根目录，执行命令 `docker build -t hub.eulix.xyz/ao-space/space-aofs:{tag} .` , 其中 `tag` 参数可以根据实际情况修改，和服务器整体运行的 docker-compose.yml 保持一致即可。

### 部署

请参考 [服务器部署](https://github.com/ao-space/ao.space/blob/dev/docs/build-and-deploy_CN.md#%E6%9C%8D%E5%8A%A1%E7%AB%AF%E9%83%A8%E7%BD%B2)。

## 本地运行

### 开发语言

- Golang 1.18+

### 依赖服务

1. Redis 5.0 +
2. PostgreSQL 11.0 +

### 服务运行所需环境变量

- SQL_HOST: 数据库的访问地址，不含端口
- SQL_PORT：数据库的访问端口
- SQL_USER： 数据库的访问账号
- SQL_PASSWORD：数据库的访问账号密码
- SQL_DATABASE：访问的数据库
- DATA_PATH: 数据存放目录
- REDIS_URL：redis的连接url
- REDIS_PASS：redis的密码
- REDIS_DB：redis的库，默认为0
- STREAM_MQ_MAX_LEN：队列长度，默认1000
- REDIS_STREAM_NAME：fileChangelogs，队列名
- GIN_MODE：debug/release(影响日志输出策略，以及是否启用swagger)

### swag 文档生成

- 安装 swag 生成工具： `go install github.com/swaggo/swag/cmd/swag@latest``
- 进入模块目录，执行命令： `swag init --parseDependency --parseInternal --o ./routers/api/docs`
  
### 运行

- 安装 golang
- 启动 Redis
- 启动 PostgreSQL
- 设置环境变量
- 本地运行服务 - 进入项目代码根目录，执行命令 `go run .`

## 贡献指南

我们非常欢迎对本项目进行贡献。以下是一些指导原则和建议，希望能够帮助您参与到项目中来。

[贡献指南](https://github.com/ao-space/ao.space/blob/dev/docs/cn/contribution-guidelines.md)

## 联系我们

- 邮箱：<developer@ao.space>
- [官方网站](https://ao.space)
- [讨论组](https://slack.ao.space)

## 感谢您的贡献

最后，感谢您对本项目的贡献。我们欢迎各种形式的贡献，包括但不限于代码贡献、问题报告、功能请求、文档编写等。我们相信在您的帮助下，本项目会变得更加完善和强大。
