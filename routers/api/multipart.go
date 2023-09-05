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
	"aofs/internal/env"
	"aofs/internal/proto"
	"aofs/repository/bpredis"
	"aofs/repository/dbutils"
	"aofs/repository/storage"
	"encoding/hex"
	"errors"
	"strconv"

	"aofs/services/multipart"
	"fmt"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
)

var rwmutex sync.RWMutex
var stor storage.MultiDiskStorager

// CreateMultipartTask
// @Summary Creating a multipart task
// @Description
// @Tags Multipart
// @Accept application/json
// @Produce application/json
// @Param userId query string true "user id"
// @Param requestId query string true "request id"
// @Param object body proto.CreateMultipartTaskReq true "params"
// @Success 200 {object} proto.Rsp{results=proto.CreateMultipartTaskRsp} "task info"
// @Router /space/v1/api/multipart/create [POST]
func CreateMultipartTask(c *gin.Context) {

	ctx := bpctx.NewCtx(c)
	rwmutex.Lock()
	defer rwmutex.Unlock()

	var req proto.CreateMultipartTaskReq
	var rsp proto.CreateMultipartTaskRsp

	if err := c.ShouldBind(&req); err != nil {
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}

	defer ctx.LogI("createMultipartTask", &req)
	if err := validate.Struct(req); err != nil {
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}


	spaceLimit := c.Query("spaceLimit")

	spaceLimitInt, _ := strconv.Atoi(spaceLimit)
	redis := bpredis.GetRedis()
	if spaceLimitInt != 0 {
		// 在线试用用户容量判断
		usedStorage, err := redis.GetInt64(bpredis.UsedSpace + strconv.Itoa(int(ctx.GetUserId())))
		if err != nil {
			usedStorage, err = dbutils.GetUsedSpaceByUser(ctx.GetUserId())
			if err != nil {
				logger.LogE().Err(err).Msg("GetUsedSpaceByUser error")
			}
		}
		if usedStorage >= int64(spaceLimitInt) {
			ctx.SendErr(proto.CodeNotEnoughSpace, errors.New("not Enough Space"))
			return
		}
	}

	param := req
	if tagLen := len(param.BETag); tagLen == 34 {
		if sizeFlag, err := hex.DecodeString(param.BETag[:2]); err != nil || sizeFlag[0] != multipart.GetSizeFlag(param.Size) {
			ctx.SendErr(proto.CodeReqParamErr, fmt.Errorf("size flag err:%v", err))
			return
		}
	} else if tagLen == 32 {
		param.BETag = hex.EncodeToString([]byte{multipart.GetSizeFlag(param.Size)}) + param.BETag
	} else {
		ctx.SendErr(proto.CodeReqParamErr, fmt.Errorf("betag err"))
		return
	}

	if len(param.FolderId) == 0 && len(param.FolderPath) == 0 {
		ctx.SendErr(proto.CodeReqParamErr, fmt.Errorf("folder param err"))
		return
	} else if len(param.FolderId) > 0 {
		if fi, err := dbutils.GetFileInfoWithUid(ctx.GetUserId(), param.FolderId); err != nil || !fi.IsDir {
			ctx.SendErr(proto.CodeReqParamErr, fmt.Errorf("folder param err"))
			return
		} else {
			param.FolderPath = fi.AbsPath()
		}
	} else {
		//循环创建目录
		if fi, err := dbutils.RecursiveCreateFolder(ctx.GetUserId(), param.FolderPath); err != nil {
			ctx.SendErr(proto.CodeReqParamErr, fmt.Errorf("failed to create folder(%s):%v", param.FolderPath, err))
			return
		} else {
			param.FolderId = fi.Id
		}
	}

	//处理秒传
	if ok, _ := stor.IsExist(env.NORMAL_BUCKET, param.BETag); ok {
		if fi, err := multipart.InsertIndex(ctx, param, false, nil); err == nil {
			rsp.RspType = proto.CREATE_MULTIPART_TASK_COMPLETE
			rsp.CompleteInfo = &fi
			ctx.SendOk(&rsp)
			return
		} else {
			ctx.SendErr(proto.CodeMultipartTaskCompleteErr, err)
			return
		}
	}

	task, err := multipart.Taskmgr.GenTask(param)
	if errors.Is(err, os.ErrExist) {
		conflict := task.GetTaskInfo()
		rsp.RspType = proto.CREATE_MULTIPART_TASK_CONFLICT
		rsp.ConflictInfo = &conflict

		ctx.SendOk(&rsp)
		return
	} else if errors.Is(err, storage.ErrEnoughSpace) {
		ctx.SendErr(proto.CodeNotEnoughSpace, err)
		return
	}

	if err != nil {
		//失败
		ctx.SendErr(proto.CodeFailedToCreateMultipartTask, err)
		return
	}
	task.UpdateLast()

	var succInfo proto.CreateMultipartTaskSuccRsp
	succInfo.UploadId = task.UploadId
	succInfo.PartSize = multipart.HASH_PART_SIZE

	rsp.RspType = proto.CREATE_MULTIPART_TASK_SUCC
	rsp.SuccInfo = &succInfo
	ctx.Append("taskInfo", task)
	ctx.SendOk(&rsp)
}

// DeleteMultipartTask
// @Summary Delete multipart task
// @Description
// @Tags Multipart
// @Accept application/json
// @Produce application/json
// @Param userId query string true "user id"
// @Param requestId query string true "request id"
// @Param object body proto.DeleteMultipartTaskReq true "params"
// @Success 200 {object} proto.Rsp ""
// @Router /space/v1/api/multipart/delete [POST]
func DeleteMultipartTask(c *gin.Context) {

	ctx := bpctx.NewCtx(c)

	rwmutex.Lock()
	defer rwmutex.Unlock()

	var req proto.DeleteMultipartTaskReq

	if err := c.ShouldBind(&req); err != nil {
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}
	defer ctx.LogI("deleteMultipartTask", &req)
	if err := validate.Struct(req); err != nil {
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}

	multipart.Taskmgr.DeleteTask(req.UploadId)
	ctx.SendOk(nil)
}

// ListMultipartTask
// @Summary Query multipart task info
// @Description
// @Tags Multipart
// @Accept application/json
// @Produce application/json
// @Param userId query string true "user id"
// @Param requestId query string true "request id"
// @Param object body proto.ListMultipartReq true "param"
// @Success 200 {object} proto.Rsp{results=proto.ListMultipartRsp} ""
// @Router /space/v1/api/multipart/list [GET]
func ListMultipartTask(c *gin.Context) {

	ctx := bpctx.NewCtx(c)

	rwmutex.RLock()
	defer rwmutex.RUnlock()

	var req proto.ListMultipartReq
	var rsp proto.ListMultipartRsp

	if err := c.ShouldBind(&req); err != nil {
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}
	defer ctx.LogI("listMultipartTask", &req)

	if err := validate.Struct(req); err != nil {
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}

	if task, err := multipart.Taskmgr.GetTask(req.UploadId); err != nil {
		ctx.SendErr(proto.CodeFailedToCreateMultipartTask, err)
		return
	} else {
		task.UpdateLast()
		rsp.UploadedParts = task.UploadedParts
		rsp.UploadingParts = task.UploadingParts
		ctx.SendOk(&rsp)
		return
	}
}

// UploadPart
// @Summary  Upload multipart data
// @Description
// @Tags Multipart
// @Accept application/json
// @Produce application/json
// @Param userId query string true "user id"
// @Param requestId query string true "request id"
// @Param object query proto.UploadPartReq true "part"
// @Param object body string true "part content"
// @Success 200 {object} proto.Rsp ""
// @Router /space/v1/api/multipart/upload [POST]
func UploadPart(c *gin.Context) {

	ctx := bpctx.NewCtx(c)

	var req proto.UploadPartReq

	if err := c.BindQuery(&req); err != nil {
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}
	defer ctx.LogI("uploadPart", &req)

	if err := validate.Struct(req); err != nil || req.Start >= req.End {
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}

	task, err := multipart.Taskmgr.GetTask(req.UploadId)
	if err != nil {
		ctx.SendErr(proto.CodeMultipartTaskNotFound, err)
		return
	}


	//进行任务上传
	if code, err := task.Upload(req.Start, req.End, c.Request.Body); err != nil {
		ctx.SendErr(code, err)
		return
	}

	task.UpdateLast()

	ctx.SendOk(nil)
}

// CompleteMultipartTask Complte multipart task
// @Summary  Complte multipart task
// @Description
// @Tags Multipart
// @Accept application/json
// @Produce application/json
// @Param userId query string true "user id"
// @Param requestId query string true "request id"
// @Param object body proto.CompleteMultipartTaskReq true "params"
// @Success 200 {object} proto.Rsp{results=proto.CompleteMultipartTaskRsp} ""
// @Router /space/v1/api/multipart/complete [POST]
func CompleteMultipartTask(c *gin.Context) {

	ctx := bpctx.NewCtx(c)

	var req proto.CompleteMultipartTaskReq
	var rsp proto.CompleteMultipartTaskRsp

	if err := c.ShouldBind(&req); err != nil {
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}
	defer ctx.LogI("completeMultipartTask", &req)

	if err := validate.Struct(req); err != nil {
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}
	task, err := multipart.Taskmgr.GetTask(req.UploadId)
	if err != nil {
		ctx.SendErr(proto.CodeMultipartTaskNotFound, err)
		return
	}
	redis := bpredis.GetRedis()
	if err := task.Complete(); err != nil {
		ctx.SendErr(proto.CodeMultipartTaskCompleteErr, err)
		return
	} else {
		if used, err := redis.GetInt64(bpredis.UsedSpace + strconv.Itoa(int(ctx.GetUserId()))); err != nil {
			redis.Set(bpredis.UsedSpace+strconv.Itoa(int(ctx.GetUserId())), task.Param.Size, 0)
		} else {
			redis.Set(bpredis.UsedSpace+strconv.Itoa(int(ctx.GetUserId())), used+task.Param.Size, 0)
		}

		rsp, err = multipart.InsertIndex(ctx, task.Param, true, task)
		if err != nil {
			ctx.SendErr(proto.CodeMultipartTaskCompleteErr, err)
			return
		}
	}

	multipart.Taskmgr.RemoveTask(req.UploadId)
	ctx.SendOk(&rsp)
}
