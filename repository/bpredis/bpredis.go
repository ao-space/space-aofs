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

package bpredis

import (
	"aofs/internal/env"
	"aofs/internal/log4bp"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

var g *bpRedis

func init() {
	Init()
}

type BpRediser interface {
	Set(key string, value interface{}, expiration time.Duration) error
	Get(key string) (int, error)
	GetInt64(key string) (int64, error)
	Incr(key string) error
	PushNotificationMsg(msg map[string]interface{})
	GetValue(key string) ([]byte, error)
	Close()
}

func GetRedis() BpRediser {
	return g
}

type bpRedis struct {
	Client  *redis.Client
	Logger  *log4bp.BpLogger
	ChanMsg chan map[string]interface{}
}

func (br *bpRedis) Close() {
	br.Client.Close()
}

func Init() error {
	var err error
	client := redis.NewClient(&redis.Options{
		Addr:     env.REDIS_URL,  // redis地址
		Password: env.REDIS_PASS, // redis密码，没有则留空
		DB:       env.REDIS_DB,   // 默认数据库，默认是0
	})
	for i := 0; i < 10; i++ {
		if _, err = client.Ping().Result(); err != nil {
			//logger.LogW().Msg(fmt.Sprintf("failed to connect to redis,err: %v", err))
			fmt.Println("failed to connect to redis", err)

			time.Sleep(time.Second)
		} else {
			//br.Logger.LogI().Msg(fmt.Sprintf("Connected to redis:%v", env.REDIS_DB))
			fmt.Println(fmt.Sprintf("Connected to redis:%v", env.REDIS_DB))
			break
		}
	}

	if err != nil {
		panic(err)
	}

	g = &bpRedis{Client: client,
		Logger:  log4bp.New("", gin.Mode()),
		ChanMsg: make(chan map[string]interface{}, 1024)}

	return nil
}
