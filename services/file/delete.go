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

package file

import (
	"aofs/internal/env"
	"aofs/internal/proto"
	"aofs/repository/dbutils"
	"aofs/services/async"
)

func MoveFilesToRecycledBin(userId proto.UserIdType, deleteIDs []string, taskList *async.TaskList) (*async.AsyncTask, *proto.BpErr) {
	//获取文件数量
	count := len(deleteIDs)

	subFiles, err := dbutils.GetSubFilesInUuids(userId, deleteIDs, []uint32{proto.TrashStatusNormal})
	if err != nil {
		return nil, &proto.BpErr{
			Code: proto.CodeFailedToOperateDB,
			Err:  err,
		}
	}

	if count < len(subFiles) {
		count = len(subFiles) + count
	}
	newTask := new(async.AsyncTask)
	newTask.Init(count)

	if count > env.ASYNC_TASK_THRESHOLD {

		taskList.Add(newTask)
	}

	code, err := dbutils.MoveFileToTrashV2(userId, deleteIDs, subFiles, count, newTask)
	if err != nil {
		return nil, &proto.BpErr{
			Code: proto.CodeFailedToOperateDB,
			Err:  err,
		}
	}

	var allFiles []string
	allFiles = append(allFiles, subFiles...)

	for _, deleteId := range deleteIDs {
		fi, _ := dbutils.GetInfoByUuid(deleteId)
		if fi != nil && !fi.IsDir {
			allFiles = append(allFiles, deleteId)
		} else if fi == nil {
			return nil, &proto.BpErr{
				Code: proto.CodeFailedToOperateDB,
				Err:  err,
			}
		}
	}

	// 历史动态：推送文件变化到网关
	go func() {
		PushChanges("file_delete", userId, allFiles)

	}()

	if code == proto.CodeCreateAsyncTaskSuccess {

		return taskList.Get(newTask.TaskId), &proto.BpErr{Code: proto.CodeCreateAsyncTaskSuccess, Err: nil}
	}
	return nil, &proto.BpErr{Code: proto.CodeOk, Err: nil}
}
