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

package storage

import (
	"encoding/json"
	"aofs/internal/env"
	"aofs/internal/log4bp"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
)

var chanMsg chan map[string]interface{}
var newClient *redis.Client

var logger = log4bp.New("", gin.Mode())

func InitClient(url string, pass string, db int) *redis.Client {
	for i := 0; i < 60; i++ {
		client := redis.NewClient(&redis.Options{
			Addr:     url,  // redis地址
			Password: pass, // redis密码，没有则留空
			DB:       db,   // 默认数据库，默认是0
		})
		_, err := client.Ping().Result()
		if err != nil {
			logger.LogW().Msg(fmt.Sprintf("failed to connect to redis,err: %v", err))
			time.Sleep(time.Second)
			//return nil, err
		} else {
			newClient = client
			logger.LogI().Msg(fmt.Sprintf("Connected to redis:%v", db))
			break
		}
	}

	if newClient == nil {
		logger.LogF().Msg("failed to connect to redis")
		//log.Println("failed to create publisher")
	} else {
		go SendMsg()
	}
	return newClient
}

func init() {
	chanMsg = make(chan map[string]interface{}, 1024)
}

func PushMsg(msg map[string]interface{}, eventName string) {
	if msg != nil && len(chanMsg) < cap(chanMsg)*3/4 {
		msg["eventName"] = eventName
		chanMsg <- msg
	}
}

func SendMsg() {
	for {
		msg, ok := <-chanMsg
		if ok {
			if newClient != nil {
				if data, err := json.Marshal(msg); err == nil {
					for i := 0; i < 60; i++ {
						pubErr := newClient.XAdd(&redis.XAddArgs{Stream: env.REDIS_STREAM_NAME,
							MaxLen: 20000,
							ID:     "*",
							Values: msg}).Err()
						if pubErr == nil {
							logger.LogI().Msg(fmt.Sprintf("successfully add msg: %v", string(data)))
							break
						} else {
							logger.LogW().Msg(fmt.Sprintf("failed to publish message,err: %v", pubErr))
						}
						if i == 59 && len(chanMsg) < cap(chanMsg)/2 {
							select {
							case chanMsg <- msg:
								logger.LogD().Interface("msg", msg).Msg("push chan again")
							default:
								break
							}
							break
						}
						time.Sleep(time.Second)
					}
				}
			}

		} else {
			break
		}
	}
}

const RedisAppPrefix = "LS-"

//
func RedisWriteUrl(shareId string, validDay uint8) error {
	err := newClient.Set(RedisAppPrefix+shareId, 0, time.Duration(validDay)*24*time.Hour).Err()
	if err != nil {
		return err
	}
	return nil
}

func RedisReadVisits(shareId string) (uv int, err error) {
	uv, err = newClient.Get(RedisAppPrefix + shareId).Int()
	if err != nil {
		return 0, err
	}
	return uv, nil
}

func RedisAddVisits(shareId string) (uv int64, err error) {
	uv, err = newClient.Incr(RedisAppPrefix + shareId).Result()
	if err != nil {
		return 0, err
	}
	return uv, nil
}



func PushStatusMsg(msg map[string]interface{}) {

	if newClient != nil {
		pubErr := newClient.XAdd(&redis.XAddArgs{Stream: "push_notification",
			MaxLen: 20000,
			ID:     "*",
			Values: msg}).Err()
		if pubErr == nil {
			logger.LogI().Msg(fmt.Sprintf("xadd backup/restore status successfully %s", msg))
			return
		} else {
			logger.LogW().Msg(fmt.Sprintf("failed to publish message,err: %v", pubErr))
		}
	}
}
