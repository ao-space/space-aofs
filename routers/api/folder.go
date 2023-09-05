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

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
)

// CreateFolders Create Folder
// @Summary Create Folder
// @Description Create Folder
// @Tags Folder
// @Accept application/json
// @Produce application/json
// @Param userId query string true "user id"
// @Param createFolderReq body proto.CreateFolderReq true "params"
// @Success 200 {object} proto.Rsp{results=proto.FileInfo} ""
// @Router /space/v1/api/folder/create [POST]
func CreateFolders(c *gin.Context) {

	ctx := bpctx.NewCtx(c)

	var crtFolderReq proto.CreateFolderReq
	defer ctx.LogI("CreateFolders", &crtFolderReq)


	validate := validator.New()
	err := validate.Struct(crtFolderReq)
	if err != nil {
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}

	// 获取请求参数
	if err := c.ShouldBind(&crtFolderReq); err != nil {
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}
	// 处理请求
	if _, err := dbutils.FileIsExist(ctx.GetUserId(), crtFolderReq.CurrentDirUuid, "", ""); err != nil {
		ctx.SendErr(proto.CodeFolderNotExist, nil)
		return
	}
	if newFolder, code, err := dbutils.CreateFolder(ctx.GetUserId(), crtFolderReq); err != nil {
		ctx.SendErr(proto.CodeFailedToCreateFolder, err)
		return
	} else if newFolder.Name == crtFolderReq.FolderName && code == 3 {
		ctx.SendOk(&newFolder)
		return
	} else if code == 2 {
		ctx.SendErr(proto.CodeFolderDepthTooLong, err)
		return
	} else if code == 1 {
		ctx.SendErr(proto.CodeFileExist, err)
		return
	}

}

// FolderInfo Get folder details
// @Summary Get folder details
// @Description Get folder details
// @Tags Folder
// @Accept application/json
// @Produce application/json
// @Param userId query string true "user id"
// @Param uuid query string  true "folder's uuid"
// @Success 200 {object} proto.Rsp{results=proto.FolderInfo} ""
// @Router /space/v1/api/folder/info [GET]
func FolderInfo(c *gin.Context) {
	ctx := bpctx.NewCtx(c)

	userId := ctx.GetUserId()


	var folderInfoReq proto.FolderInfoReq

	// 获取请求参数
	if err := c.ShouldBind(&folderInfoReq); err != nil {
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}

	//计算文件夹大小
	if fs, err := dbutils.CalculateFolderSizeWithUuid(userId, folderInfoReq.FolderUuid); err != nil {
		ctx.LogE().Err(err).Msg("failed to Calculate FolderSize ")
		return
	} else {
		logger.LogI().Int64("folderSize", fs).Msg("Calculate Folder Size")
	}
	//返回文件夹信息
	if fi, err := dbutils.GetFolderInfoByUuid(folderInfoReq.FolderUuid); err != nil {
		ctx.SendErr(proto.CodeFileNotExist, err)
		return
	} else {
		ctx.SendOk(&fi)
		return
	}

}
