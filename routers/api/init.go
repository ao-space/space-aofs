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

package api

//此文件主要完成系统初始化和用户初始化
import (
	"aofs/internal/env"
	"aofs/internal/log4bp"
	"aofs/internal/proto"
	"aofs/repository/dbutils"
	"aofs/repository/storage"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/google/uuid"
)

var logger = log4bp.New("", gin.Mode())

var validate = validator.New()

//此文件初始化系统readme文件到根目录
func InitReadme(userId proto.UserIdType) {
	trans, err := dbutils.NewTransProducter().New()
	if err != nil {
		logger.LogE().Err(err).Msg("failed to get trans")
		os.Exit(1)
	}

	defer trans.Commit()

	v, _ := trans.GetSetting(fmt.Sprintf("InitReadme-%v", userId))
	if v == "ok" {
		return
	}
	//检查是否已经存在readme.txt, *** 后续要保存记录，避免用户删除后，又重新创建。
	if exists, err := dbutils.IsExistByPath(userId, "/", "说明.pdf"); err == nil && exists {

		logger.LogD().Msg("说明.pdf is exist")
		return
	} else {

		logger.LogD().Msg("init 说明.pdf")
	}

	//如果文件不存在，则初始化一个
	{
		data, err := ioutil.ReadFile("/tmp/说明.pdf")
		if err != nil {
			logger.LogE().Err(err).Msg("failed to read 说明.pdf")
		}

		md5data := md5.Sum(data)
		readme := proto.FileInfo{
			FileInfoPub: proto.FileInfoPub{
				Id:            uuid.New().String(),
				IsDir:         false,
				Name:          "说明.pdf",
				Path:          "/",
				BETag:         hex.EncodeToString(md5data[:]),
				CreateTime:    time.Now().UnixNano() / 1e6,
				ModifyTime:    time.Now().UnixNano() / 1e6,
				OperationTime: time.Now().UnixNano() / 1e6,
				Size:          int64(len(data)),
				FileCount:     0,
				Category:      "document",
				Mime:          "application/pdf",
			},
			UserId:        userId,
			Tags:          "",
			Executable:    false,
			Version:       0,
			BucketName:    env.NORMAL_BUCKET,
			TransactionId: 0,
		}


		if err := stor.Put(env.NORMAL_BUCKET, readme.BETag, bytes.NewReader(data), int64(len(data))); err == nil {
			if err := dbutils.AddFile(readme); err == nil {
				logger.LogD().Msg("init 说明.pdf succ")

				trans.SetSetting(fmt.Sprintf("InitReadme-%v", userId), "ok")
			} else {
				logger.LogE().Err(err).Msg("init readme.pdf. failed to insert db.")
			}

		}
	}
}

func initUser(userId proto.UserIdType) error {
	dbutils.InitUser(userId)
	InitReadme(userId)
	logger.LogI().Msg(fmt.Sprintf("Init User:%v", userId))
	return nil
}

func Init() {
	stor = storage.GetStor()
	initUser(1)
}
