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
	"aofs/repository/dbutils"
	"aofs/services/file"
	"aofs/services/recycled"
	"github.com/gin-gonic/gin"
)

// ListRecycled Query the recycle bin file list
// @Summary Query the recycle bin file list
// @Description Query the recycle bin file list
// @Tags Recycled
// @Accept application/json
// @Produce application/json
// @Param userId query string true "user id"
// @Param page query int false "page index，default: 1"
// @Param pageSize query int false "page size，default: 10"
// @Success 200 {object} proto.Rsp{results=proto.GetListRspData} ""
// @Router /space/v1/api/recycled/list [GET]
func ListRecycled(c *gin.Context) {

	ctx := bpctx.NewCtx(c)

	var pageInfo proto.PageInfo
	var rspData proto.GetListRspData

	userId := ctx.GetUserId()
	//获取请求参数
	if err := c.ShouldBind(&pageInfo); err != nil {
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}

	if pageInfo.Page == 0 {
		pageInfo.Page = 1
	}
	if pageInfo.PageSize == 0 {
		pageInfo.PageSize = 10
	}

	recycleList, err := dbutils.GetRecycledList(userId, pageInfo.Page, "desc", pageInfo.PageSize)
	if err != nil {
		ctx.SendErr(proto.CodeFileNotExist, err)
		return
	} else {
		rspData.List = recycleList.ToPubLst()
		rspData.PageInfo.PageInfo = pageInfo
		rspData.PageInfo.TotalPage, rspData.PageInfo.FileCount, err = dbutils.PageTotal(userId, "", "", pageInfo.PageSize, true)
		if err != nil {
			ctx.SendErr(proto.CodeFailedToOperateDB, err)
			return
		}
		ctx.SendOk(&rspData)
		return
	}

}

// RestoreRecycled Restore files from recycle bin
// @Summary Restore files from recycle bin
// @Description Restore files from recycle bin
// @Tags Recycled
// @Accept application/json
// @Produce application/json
// @Param userId query int true "user id"
// @Param restoreFilesReq body proto.RestoreRecycledReq true "params"
// @Success 200 {object} proto.Rsp "success"
// @Success 201 {object} proto.Rsp{results=async.AsyncTask} "asynchronous task"
// @Router /space/v1/api/recycled/restore [POST]
func RestoreRecycled(c *gin.Context) {

	ctx := bpctx.NewCtx(c)
	var restoreReq proto.RestoreRecycledReq

	//获取请求参数
	if err := c.ShouldBind(&restoreReq); err != nil {
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}
	defer ctx.LogI("restoreRecycled", restoreReq)
	//处理请求
	err := validate.Struct(restoreReq)
	if err != nil {
		ctx.LogE().Err(err).Msg("validate failed")
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}

	taskInfo, bpErr := file.RestoreFilesFromRecycledBin(ctx.GetUserId(), restoreReq.RecycledUuids, taskList)
	if taskInfo == nil && bpErr.Code == proto.CodeOk {
		ctx.SendOk(nil)
		return
	} else if taskInfo != nil && bpErr.Code == proto.CodeCreateAsyncTaskSuccess {
		ctx.SendRsp(taskInfo, bpErr)
		return
	} else {
		ctx.SendErr(bpErr.Code, bpErr.Err)
		return
	}

}

// ClearRecycled Clean out the recycle bin
// @Summary Clean out the recycle bin
// @Description Clean out the recycle bin
// @Tags Recycled
// @Accept application/json
// @Produce application/json
// @Param userId query int true "user id"
// @Param uuids body proto.RecycledPhyDeleteReq true "params"
// @Success 200 {object} proto.Rsp{} ""
// @Router /space/v1/api/recycled/clear [POST]
func ClearRecycled(c *gin.Context) {
	ctx := bpctx.NewCtx(c)
	//对回收状态为1和4的，全部修改为2
	var req proto.RecycledPhyDeleteReq
	//var rsp proto.Rsp
	//rsp.SetContext(c)

	if err := c.ShouldBind(&req); err != nil {
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}
	affect, err := dbutils.RecycledFromLogicToPhy(ctx.GetUserId(), req.Uuids)

	defer ctx.LogI("clearRecycled", req)

	if err != nil {
		logger.LogE().Err(err).Msg("clear recycled failed")
		ctx.SendErr(proto.CodeFailedToOperateDB, err)
	} else {
		logger.LogI().Interface("uuids", req.Uuids).Interface("affect:", affect).Msg("clearRecycled")
		ctx.SendOk(nil)
	}
	recycled.DoClearRecycledTask()
}
