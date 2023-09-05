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
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

func init() {
	if g != nil {
		go g.SendMessage()
	}
}

const (
	PreviewQueue      = "fileChangelogs"
	NotificationQueue = "push_notification"
)

func (br *bpRedis) InsertChan(msg map[string]interface{}, eventName string) {
	if msg != nil && len(br.ChanMsg) < cap(br.ChanMsg)*3/4 {
		msg["eventName"] = eventName
		br.ChanMsg <- msg
	}
}

func (br *bpRedis) SendMessage() {
	br.Logger.LogI().Msg("start wait message ...")
	for {
		msg, ok := <-br.ChanMsg
		if ok {
			if br.Client != nil {
				if data, err := json.Marshal(msg); err == nil {
					for i := 0; i < 60; i++ {
						pubErr := br.Client.XAdd(&redis.XAddArgs{Stream: PreviewQueue,
							MaxLen: 20000,
							ID:     "*",
							Values: msg}).Err()
						if pubErr == nil {
							br.Logger.LogI().Msg(fmt.Sprintf("successfully add msg: %v", string(data)))
							break
						} else {
							br.Logger.LogW().Msg(fmt.Sprintf("failed to publish message,err: %v", pubErr))
						}
						if i == 59 && len(br.ChanMsg) < cap(br.ChanMsg)/2 {
							select {
							case br.ChanMsg <- msg:
								br.Logger.LogD().Interface("msg", msg).Msg("push chan again")
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


func (br *bpRedis) PushNotificationMsg(msg map[string]interface{}) {
	//dataBytes,_ := json.Marshal(msg["data"])
	//if infoBytes, err := base64.StdEncoding.DecodeString(string(dataBytes));err == nil  {
	//	json.Unmarshal(infoBytes,proto.)
	//}
	if br.Client != nil {
		pubErr := br.Client.XAdd(&redis.XAddArgs{Stream: NotificationQueue,
			MaxLen: 20000,
			ID:     "*",
			Values: msg}).Err()
		if pubErr == nil {
			br.Logger.LogI().Msg(fmt.Sprintf("xadd message successfully %s", msg))
			return
		} else {
			br.Logger.LogW().Msg(fmt.Sprintf("failed to publish message,err: %v", pubErr))
		}
		//time.Sleep(time.Second)
	}
}
