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

package recycled

import (
	"aofs/internal/env"
	"aofs/internal/log4bp"
	"aofs/internal/proto"
	"aofs/repository/bpredis"
	"aofs/repository/dbutils"
	"aofs/repository/storage"
	"fmt"
	"strconv"

	"time"

	"github.com/gin-gonic/gin"
)

var stor storage.MultiDiskStorager
var logger = log4bp.New("", gin.Mode())

func dispatchClearTask() {
	if len(chClearRecycled) < 1 {
		chClearRecycled <- time.Now().String()
	}
}

var chClearRecycled chan string

func Init() {
	stor = storage.GetStor()
	chClearRecycled = make(chan string, 3)
	chClearRecycled <- "init"

	go doClearRecycled()
	go timerClearRecycled()
}

func timerClearRecycled() {
	for {
		time.Sleep(time.Hour)
		dispatchClearTask()
	}
}

func doClearRecycled() {
	for task := range chClearRecycled {
		logger.LogD().Interface("doClearRecycled", task).Msg("deal clear recycled task")
		DoClearRecycledTask()
	}
}

func doClearRecycledFolder(dir proto.FileInfo) error {
	//直接删除记录
	if _, err := dbutils.DeleteByUuid(dir.Id); err != nil {
		logger.LogE().Err(err).Msg(fmt.Sprintf("failed to remove dir: %v,%v", dir.Id, dir.Name))
		dbutils.RecycledFromPhyToException(dir.Id) //放到异常队列，后续重试处理
	} else {
		logger.LogI().Msg(fmt.Sprintf("success to remove dir: %v,%v", dir.Id, dir.Name))
	}

	return nil
}

func DoClearRecycledFile(file proto.FileInfo) error {
	//查询md5是否存在共用
	if sharCnt, err := dbutils.GetSharedCntByBEtag(file.BETag); err != nil {
		logger.LogE().Err(err).Msg("Failed to GetSharedCntByMd5sum")
		return err
	} else if sharCnt <= 1 {
		stor.Del(env.NORMAL_BUCKET, file.BETag)
		redis := bpredis.GetRedis()
		if used, err := redis.GetInt64(bpredis.UsedSpace + strconv.Itoa(int(file.UserId))); err != nil {
			usedSpace, _ := dbutils.GetUsedSpaceByUser(file.UserId)
			redis.Set(bpredis.UsedSpace+strconv.Itoa(int(file.UserId)), usedSpace, 0)
		} else {
			redis.Set(bpredis.UsedSpace+strconv.Itoa(int(file.UserId)), used-file.Size, 0)
		}
	} else {
		logger.LogW().Msg("there is a same betag file,cancel clear real file")
	}

	//从数据库中删除记录
	if affect, err := dbutils.DeleteByUuid(file.Id); err != nil {
		logger.LogE().Err(err).Msg(fmt.Sprintf("failed to remove file:%v,%v", file.Id, file.Name))
		dbutils.RecycledFromPhyToException(file.Id) //放到异常队列，后续重试处理
	} else {
		logger.LogI().Msg(fmt.Sprintf("success to remove file: %v,%v,%v", file.Id, file.Name, affect))
	}

	return nil
}

func doClearRecycledList(files []proto.FileInfo, folderCnt *int, fileCnt *int) error {
	//执行删除动作
	for _, file := range files {
		if file.IsDir {
			(*folderCnt)++
			if err := doClearRecycledFolder(file); err != nil {
				return err
			}

		} else {
			(*fileCnt)++
			if err := DoClearRecycledFile(file); err != nil {
				return err
			}
		}
	}

	return nil
}
func DoClearRecycledTask() {

	folderCnt := 0
	fileCnt := 0
	defer func() {
		logger.LogD().Msg("finish to clear")
	}()

	for {
		if files, err := dbutils.GetRecycledPhyDeletedList(0, 1024); err != nil {
			logger.LogE().Err(err).Msg("failed to get recycle list")
			return
		} else if len(files) == 0 {
			logger.LogD().Interface("folder cnt", folderCnt).Interface("file cnt", fileCnt).Msg("clear recycle")
			return
		} else {
			if err := doClearRecycledList(files, &folderCnt, &fileCnt); err != nil {
				return
			}
		}
	}

}
