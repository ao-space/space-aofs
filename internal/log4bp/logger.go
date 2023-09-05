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

package log4bp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

//var logger = New("", gin.Mode())

type BpLogger struct {
	*zerolog.Logger
	tid string
}

// type rspInterface interface {
// 	GetCode() uint32
// 	GetMsg() string
// 	GetBody() interface{}
// }

type SharedBoxInfo struct {
	BoxUuid string `json:"boxUuid"` // 盒子端唯一id.
	Btid    string `json:"btid"`
}

func GetBoxUuid(filename string) (sharedBoxInfo SharedBoxInfo, err error) {
	if _, err := os.Stat(filename); err != nil {
		return sharedBoxInfo, err
	}

	if f, err := os.Open(filename); err == nil {
		defer f.Close()
		//r := io.Reader(f)

		if err = json.NewDecoder(f).Decode(&sharedBoxInfo); err != nil {
			return sharedBoxInfo, fmt.Errorf("GetBoxUuid Failed:%v", err)
		}
		return sharedBoxInfo, nil
	} else {
		return sharedBoxInfo, err
	}

}

func (l *BpLogger) LogI() *zerolog.Event {
	return l.Logger.Info().Timestamp().Caller(1).Str("tid", l.tid)
}
func (l *BpLogger) LogD() *zerolog.Event {
	return l.Logger.Debug().Timestamp().Caller(1).Str("tid", l.tid)
}
func (l *BpLogger) LogW() *zerolog.Event {
	return l.Logger.Warn().Timestamp().Caller(1).Str("tid", l.tid)
}
func (l *BpLogger) LogE() *zerolog.Event {
	return l.Logger.Error().Timestamp().Caller(1).Str("tid", l.tid)
}
func (l *BpLogger) LogF() *zerolog.Event {
	return l.Logger.Fatal().Timestamp().Caller(1).Str("tid", l.tid)
}
func (l *BpLogger) LogP() *zerolog.Event {
	return l.Logger.Panic().Timestamp().Caller(1).Str("tid", l.tid)
}

func (l *BpLogger) SetTid(tid string) {
	l.tid = tid
}

var Logger zerolog.Logger

func init() {
	zerolog.CallerMarshalFunc = func(file string, n int) string {
		_, file = filepath.Split(file)
		return fmt.Sprintf("%s:%d", file, n)
	}
	zerolog.CallerFieldName = "file"
	zerolog.LevelFieldName = "level"
	zerolog.MessageFieldName = "message"
	zerolog.TimestampFieldName = "time"

	Logger = zerolog.New(os.Stdout)
}

func New(tid string, mode string) *BpLogger {

	if len(tid) == 0 {
		tid = uuid.New().String()[:13]
	}
	if mode == "debug" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	return &BpLogger{&Logger, tid}
}
