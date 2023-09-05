// Copyright (c) 2022 Institute of Software, Chinese Academy of Sciences (ISCAS)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

// @title aofs apis
// @version 1.0
// @description This is AO.space aofs OpenAPI reference document.
// @termsOfService http://swagger.io/terms/

// @contact.name AO.space
// @contact.url https://ao.space/
// @contact.email service@ao.space

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath /

import (
	"aofs/internal/env"
	"aofs/internal/log4bp"
	"aofs/repository/dbutils"
	"aofs/repository/storage"
	"aofs/routers/api"
	_ "aofs/routers/api/docs"
	"aofs/routers/routers"
	"aofs/services/multipart"
	"aofs/services/recycled"
	"fmt"

	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"gorm.io/gorm"
)

type FileAPI struct {
	Route    *gin.Engine
	Logger   *log4bp.BpLogger
	RedisCli *redis.Client
	Stor     storage.MultiDiskStorager
	DB       *gorm.DB
}

var fileapi *FileAPI

func NewFileAPI() *FileAPI {
	//stor := storage.GetStor()
	return &FileAPI{
		Route:    routers.InitRoute(),
		Logger:   log4bp.New("", gin.Mode()),
		RedisCli: storage.InitClient(env.REDIS_URL, env.REDIS_PASS, env.REDIS_DB),
		Stor:     storage.GetStor(),
		DB:       dbutils.GetFileDB(),
	}
}

func Init() {
	dbutils.Init()

	if err := storage.Init(dbutils.NewBETagIndexer()); err != nil {
		fileapi.Logger.LogE().Err(err).Msg("failed to storage.Init")
		os.Exit(2)
	}


	api.Init()
	recycled.Init() //回收站初始化
	multipart.Init()
}

func main() {
	fileapi = NewFileAPI()
	fileapi.Route.Use(gin.Recovery())
	fileapi.Route.MaxMultipartMemory = 100 * 1024 * 1024
	Init()

	go func() {
		err := fileapi.Route.Run(":2001")
		if err != nil {
			fmt.Println("run err:", err)
		}
	}()

	//等待中断信号，以优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	// 可以捕捉除了kill-9的所有中断信号
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt)
	<-quit


	fileapi.Logger.LogI().Msg("receive exit signal ,exit....")


}
