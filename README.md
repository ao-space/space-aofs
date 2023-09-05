# space-aofs

English | [简体中文](./README_cn.md)

space-aofs (AO.space File Service) is an internal module that provides file services for the AO.space server. It works in conjunction with the space-gateway and space-nginx modules to provide file access interfaces, including file listing, file upload, file download, and file management APIs.

## Build And Run

### Compilation environment preparation

Please install the latest version of Docker.

### Source code download

Please follow the steps below to download the source code:

- Create a local directory, run cmd: `mkdir space-aofs`
- Enter the local directory, run cmd: `cd ./space-aofs`
- Run cmd: `git clone git@github.com:ao-space/space-aofs.git .`

### Container image building

Run cmd: `docker build -t hub.eulix.xyz/ao-space/space-aofs:{tag} .` , Where the `tag` parameter can be modified according to the actual situation.。

### Deploy

Please refer to [server deploy](https://github.com/ao-space/ao.space/blob/dev/docs/build-and-deploy.md#server-deploy).

## Local running

### Development language

- Golang 1.18+

### Dependent services

1. Redis 5.0 +
2. PostgreSQL 11.0 +

### Environment variables

- SQL_HOST: PostgreSQL server host
- SQL_PORT：PostgreSQL server port
- SQL_USER： PostgreSQL access account
- SQL_PASSWORD：PostgreSQL access password
- SQL_DATABASE：PostgreSQL database name
- DATA_PATH: path for data saving
- REDIS_URL：redis server url
- REDIS_PASS：redis server password
- STREAM_MQ_MAX_LEN：length for message queue，default: 1000
- REDIS_STREAM_NAME：message queue, default: fileChangelogs
- GIN_MODE：debug/release

### Swag document generation

- Install swag ： `go install github.com/swaggo/swag/cmd/swag@latest``
- Run cmd： `swag init --parseDependency --parseInternal --o ./routers/api/docs`
  
### Run

- Install golang
- Start Redis
- Start PostgreSQL
- Set up the environment variables
- Run cmd: `go run .`

## Contribution Guidelines

Contributions to this project are very welcome. Here are some guidelines and suggestions to help you get involved in the project.

[Contribution Guidelines](https://github.com/ao-space/ao.space/blob/dev/docs/en/contribution-guidelines.md)

## Contact us

- Email: <developer@ao.space>
- [Official Website](https://ao.space)
- [Discussion group](https://slack.ao.space)

## Thanks for your contribution

Finally, thank you for your contribution to this project. We welcome contributions in all forms, including but not limited to code contributions, issue reports, feature requests, documentation writing, etc. We believe that with your help, this project will become more perfect and stronger.
