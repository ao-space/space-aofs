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

import (
	"aofs/internal/bpctx"
	"aofs/internal/proto"
	"aofs/services/async"

	"github.com/gin-gonic/gin"
)

var taskList = async.NewTaskList()

// GetAsyncTaskInfo @Summary Query the status of an asynchronous task
// @Description Query the status of an asynchronous task
// @Tags Async
// @Accept application/json
// @Produce application/json
// @Param userId query int true "user id"
// @Param taskId query string true "task id"
// @Success 200 {object} proto.Rsp{results=async.AsyncTask} "response"
// @Router /space/v1/api/async/task [GET]
func GetAsyncTaskInfo(c *gin.Context) {
	ctx := bpctx.NewCtx(c)

	var taskReq proto.AsyncStaskInfoReq

	if err := c.ShouldBind(&taskReq); err != nil {
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}

	taskInfo, err := taskList.GetTaskStatus(taskReq.TaskId)
	if err != nil {
		ctx.SendErr(proto.CodeGetAsyncTaskInfoFailed, err)
		return
	}
	defer func() {
		ctx.LogI("async task", taskReq)
		//time.Sleep(time.Second)
		if taskInfo.TaskStatus == async.AsyncTaskStatusSuccess || taskInfo.TaskStatus == async.AsyncTaskStatusFailed {
			taskList.Remove(taskInfo.TaskId)
		}
	}()

	ctx.SendOk(taskInfo)
}
