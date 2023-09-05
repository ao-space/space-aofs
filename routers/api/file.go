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
	"aofs/repository/dbutils"
	"aofs/services/file"
	"errors"
	"os"
	"path/filepath"
	"time"

	"aofs/internal/proto"
	"fmt"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	_ "github.com/swaggo/swag/example/celler/httputil"
)

// MoveFile Move file/folder
// @Summary Move file/folder
// @Description Move file/folder
// @Tags File
// @Accept application/json
// @Produce application/json
// @Param userId query string true "user id"
// @Param moveFilesReq body proto.MoveFileReq. true "params"
// @Success 200 {object} proto.Rsp{results=proto.DbAffect} ""
// @Router /space/v1/api/file/move [POST]
func MoveFile(c *gin.Context) {

	ctx := bpctx.NewCtx(c)

	var moveFileReq proto.MoveFileReq
	var moveAffect proto.DbAffect
	//var rsp proto.Rsp
	// rsp.SetContext(c)
	userId := ctx.GetUserId()
	//获取请求参数
	if err := c.ShouldBind(&moveFileReq); err != nil {
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}

	//请求参数验证
	err := validate.Struct(moveFileReq)
	if err != nil {
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}
	defer ctx.LogI("moveFiles", moveFileReq)
	// 移动的uuid 和目录uuid 不能相同
	for _, value := range moveFileReq.Id {
		if value == moveFileReq.DestPath {
			ctx.SendErr(proto.CodeCopyIdError, nil)
			return
		}
	}
	//开始处理请求
	for i, id := range moveFileReq.Id {
		/*判断文件是否存在*/
		if affect, err := dbutils.FileIsExist(userId, id, "", ""); affect == 1 {
			if affectRow, err := dbutils.MoveFiles(userId, id, moveFileReq.DestPath); err != nil {
				ctx.LogE().Err(err).Msg(fmt.Sprintf("Move File %v Failed ", id))
				//ctx.SendErr(proto.CodeFailedToOperateDB, err)
				//return
			} else if affectRow != 0 {
				moveAffect.AffectRows = +uint32(affectRow)
			}
		} else {
			ctx.LogE().Err(err).Msg(fmt.Sprintf("File %v Not Exist", id))
		}
		if i == len(moveFileReq.Id)-1 {
			if moveAffect.AffectRows > 0 {
				ctx.SendOk(&moveAffect)
			} else if errors.Is(err, errors.New("beyond 20 layers")) {
				ctx.SendErr(proto.CodeFolderDepthTooLong, err)
			} else {
				ctx.SendErr(proto.CodeFileNotExist, err)
			}
		}
	}

	moveId := moveFileReq.Id[0]
	pathRecord, nameRecord := dbutils.GetPathByUuid(moveId)

	// 计算文件夹内文件数
	if fc, err := dbutils.CalculateFileCount(userId, pathRecord, nameRecord); err != nil {
		logger.LogE().Err(err).Msg("calculate files failed !")
	} else {
		if len(pathRecord) == 0 {
			logger.LogD().Msg(fmt.Sprintf("%s： %d files\r\n", pathRecord, fc))
		} else {
			logger.LogD().Msg(fmt.Sprintf("/： %d files\r\n", fc))
		}
	}
}

// ListFiles Get file list
// @Summary Get file list
// @Description Get file list
// @Tags File
// @Accept application/json
// @Produce application/json
// @Param userId query int true "user id"
// @Param uuid query string false "folder uuid"
// @Param isDir query bool false "Whether to filter folders"
// @Param page query int false "page，default:1"
// @Param pageSize query int false "page size，default:10"
// @Param orderBy query string false "Sort. The default is reverse order"
// @Param category query string false  "file classification, field value: document，video，picture or other; If there is no field, all are included"
// @Success 200 {object} proto.Rsp{results=proto.GetListRspData}
// @Router /space/v1/api/file/list [get]
func ListFiles(c *gin.Context) {

	ctx := bpctx.NewCtx(c)

	var req proto.GetListReq
	var rspData proto.GetListRspData
	//var rsp proto.Rsp
	//var fileinfo proto.FileInfo
	//rsp.SetContext(c)

	userId := ctx.GetUserId()

	//获取请求参数
	if err := c.ShouldBind(&req); err != nil {
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}


	if req.PageInfo.Page == 0 {
		req.PageInfo.Page = 1
	}
	if req.PageInfo.PageSize == 0 {
		req.PageInfo.PageSize = 10
	}
	//
	if req.OrderBy == "" {
		req.OrderBy = "is_dir desc,operation_time DESC"
	}

	// category 参数为空则返回全部文件列表
	// category 参数不为空则返回分类文件列表（视频，文档，图片）
	if req.Category == "" {
		if req.Uuid == "" {
			rootList, err := dbutils.GetRootList(userId, req.IsDir, req.PageInfo.Page, req.OrderBy, req.PageInfo.PageSize)
			if err != nil {
				ctx.SendErr(proto.CodeFailedToOperateDB, err)
				return
			}
			rspData.List = rootList.ToPubLst()
			rspData.PageInfo.PageInfo = req.PageInfo
			rspData.PageInfo.TotalPage, rspData.PageInfo.FileCount, err = dbutils.PageTotal(userId, req.Uuid, req.Category, req.PageInfo.PageSize, false)
			if err != nil {
				ctx.SendErr(proto.CodeFolderNotExist, err)
				return
			}
		} else if affect, err := dbutils.FileIsExist(userId, req.Uuid, "", ""); affect == 0 {
			ctx.SendErr(proto.CodeFolderNotExist, err)
			return
		} else {
			fileList, err := dbutils.GetFileList(ctx.GetUserId(), req.IsDir, req.Uuid, req.PageInfo.Page, req.OrderBy, "", int(req.PageInfo.PageSize))
			if err != nil {
				ctx.SendErr(proto.CodeFailedToOperateDB, err)
				return
			}
			rspData.List = fileList.ToPubLst()
			rspData.PageInfo.PageInfo = req.PageInfo
			rspData.PageInfo.TotalPage, rspData.PageInfo.FileCount, err = dbutils.PageTotal(userId, req.Uuid, req.Category, req.PageInfo.PageSize, false)
			if err != nil {
				ctx.SendErr(proto.CodeFolderNotExist, err)
				return
			}

		}
	} else {
		fileList, err := dbutils.GetFileList(ctx.GetUserId(), req.IsDir, "", req.PageInfo.Page, req.OrderBy, req.Category, int(req.PageInfo.PageSize))
		if err != nil {
			ctx.SendErr(proto.CodeFailedToOperateDB, err)
			return
		}
		rspData.List = fileList.ToPubLst()
		rspData.PageInfo.PageInfo = req.PageInfo
		rspData.PageInfo.TotalPage, rspData.PageInfo.FileCount, err = dbutils.PageTotal(userId, req.Uuid, req.Category, req.PageInfo.PageSize, false)
		if err != nil {
			ctx.SendErr(proto.CodeFolderNotExist, err)
			return
		}
	}

	//发送处理结果
	ctx.SendOk(&rspData)
}

// ModifyFile
// @Summary Modify file/folder name
// @Description Modify file/folder name
// @Tags File
// @Accept application/json
// @Produce application/json
// @Param userId query string true "user id"
// @Param ModifyFileReq body proto.ModifyFileReq true "request info"
// @Success 200 {object} proto.Rsp{results=proto.DbAffect} ""
// @Router /space/v1/api/file/rename [POST]
func ModifyFile(c *gin.Context) {

	ctx := bpctx.NewCtx(c)

	var modifyFileReq proto.ModifyFileReq
	var modifyAffect proto.DbAffect

	userId := ctx.GetUserId()
	// 获取请求参数
	if err := c.ShouldBind(&modifyFileReq); err != nil {
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}

	err := validate.Struct(modifyFileReq)
	if err != nil {
		ctx.LogE().Err(err).Msg("validate failed")
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}
	defer ctx.LogI("modifyFile", modifyFileReq)
	// 处理请求
	if affectRow, err := dbutils.RenameFiles(userId, modifyFileReq.Id, modifyFileReq.NewFileName); err != nil {
		ctx.SendErr(proto.CodeFailedToOperateDB, err)
		return
	} else if affectRow != 0 {
		modifyAffect.AffectRows = uint32(affectRow)
		ctx.SendOk(&modifyAffect)
		return
	}
}

// TrashFiles Delete files/folders to Recycle bin
// @Summary Delete files/folders to Recycle bin
// @Description Delete files/folders to Recycle bin
// @Tags File
// @Accept application/json
// @Produce application/json
// @Param userId query string true "user id"
// @Param DeleteId body proto.DeleteFileReq true "params"
// @Success 200 {object} proto.Rsp ""
// @Success 201 {object} proto.Rsp{results=async.AsyncTask} ""
// @Failure 400,404 {object} httputil.HTTPError
// @Router /space/v1/api/file/delete [POST]
func TrashFiles(c *gin.Context) {

	ctx := bpctx.NewCtx(c)

	var deleteReq proto.DeleteFileReq


	userId := ctx.GetUserId()
	//获取请求参数

	if err := c.ShouldBind(&deleteReq); err != nil {
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}
	defer ctx.LogI("deleteFiles", deleteReq)

	err := validate.Struct(deleteReq)
	if err != nil {
		ctx.LogE().Err(err).Msg("validate failed")
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}

	//改造
	// 第一步：改造MoveFileToTrashV2，让它可以返回code，如果处理量小则直接返回code-200
	// 处理量大的话，返回201

	taskInfo, bpErr := file.MoveFilesToRecycledBin(userId, deleteReq.DeleteIds, taskList)
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

// CopyFiles Batch copy files
// @Summary Batch copy files
// @Description Batch copy files
// @Tags File
// @Accept application/json
// @Produce application/json
// @Param userId query string true "user id"
// @Param copyFilesReq body proto.CopyFileReq true "params"
// @Success 200 {object} proto.Rsp{results=proto.CopyRsp} ""
// @Router /space/v1/api/file/copy [POST]
func CopyFiles(c *gin.Context) {
	ctx := bpctx.NewCtx(c)

	var copyFileReq proto.CopyFileReq
	var copyRsp proto.CopyRsp

	userId := ctx.GetUserId()

	//获取请求参数
	if err := c.ShouldBind(&copyFileReq); err != nil {
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}
	defer ctx.LogI("copyFiles", copyFileReq)

	err := validate.Struct(copyFileReq)
	if err != nil {
		ctx.LogE().Err(err).Msg("validate failed")
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}
	for _, value := range copyFileReq.Ids {
		if value == copyFileReq.DestId {
			ctx.SendErr(proto.CodeCopyIdError, nil)
			return
		}
	}

	if affect, err := dbutils.FileIsExist(userId, copyFileReq.DestId, "", ""); affect == 1 {
		if affectRow, newAndOldId, err := dbutils.CopyFile(userId, copyFileReq); err != nil {
			if errors.Is(err, errors.New("beyond 20 layers")) {
				ctx.SendErr(proto.CodeFolderDepthTooLong, err)
				return
			}
			ctx.SendErr(proto.CodeFailedToOperateDB, err)
			return
		} else {
			copyRsp.AffectRows = uint32(affectRow)
			copyRsp.Data = newAndOldId
			ctx.SendOk(&copyRsp)
		}

	} else {
		ctx.SendErr(proto.CodeFolderNotExist, err)
		return
	}

	copyId := copyFileReq.Ids[0]
	pathRecord, nameRecord := dbutils.GetPathByUuid(copyId)
	// 计算文件夹内文件数
	if fc, err := dbutils.CalculateFileCount(userId, pathRecord, nameRecord); err != nil {
		//log.Println(pathRecord+": calculate files failed !", err)
		logger.LogE().Err(err).Msg("calculate files failed !")
	} else {
		if len(pathRecord) == 0 {
			logger.LogD().Msg(fmt.Sprintf("%s： %d files\r\n", pathRecord, fc))
		} else {
			logger.LogD().Msg(fmt.Sprintf("/： %d files\r\n", fc))
		}
	}

}

// SearchFiles Search files
// @Summary Search files
// @Description Search files
// @Tags File
// @Accept application/json
// @Produce application/json
// @Param userId query string true "user id"
// @Param uuid query string false "folder's uuid，default: /"
// @Param name query string true "filename"
// @Param category query string false "file type"
// @Param page query int false "page, default:1"
// @Param pageSize query int false "page size，default:10"
// @Param orderBy query string false "sort type, default value is in reverse order of change time"
// @Success 200 {object} proto.Rsp{results=proto.GetListRspData} ""
// @Router /space/v1/api/file/search [GET]
func SearchFiles(c *gin.Context) {

	ctx := bpctx.NewCtx(c)

	var searchReq proto.SearchReq
	var rspData proto.GetListRspData

	userId := ctx.GetUserId()

	//获取请求参数
	if err := c.ShouldBind(&searchReq); err != nil {
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}
	defer ctx.LogI("searchFiles", searchReq)

	//开始处理请求
	if searchReq.ObjectName != "/" /*根据实际情况*/ {


		if searchReq.PageInfo.Page == 0 {
			searchReq.PageInfo.Page = 1
		}
		if searchReq.PageInfo.PageSize == 0 {
			searchReq.PageInfo.PageSize = 10
		}
		if searchReq.OrderBy == "" {
			searchReq.OrderBy = "is_dir DESC"
		}

		allMatchingFiles, err := dbutils.SearchFileByName(userId, searchReq.Uuid, searchReq.ObjectName, searchReq.Category, searchReq.OrderBy, searchReq.PageInfo.Page, searchReq.PageInfo.PageSize)
		if err != nil {
			ctx.SendErr(proto.CodeFileNotExist, err)
			return
		}
		rspData.List = allMatchingFiles.ToPubLst()
		rspData.PageInfo.PageInfo = searchReq.PageInfo
		rspData.PageInfo.TotalPage, rspData.PageInfo.FileCount, err = dbutils.SearchPageTotal(userId, searchReq.Uuid, searchReq.ObjectName, searchReq.Category, searchReq.OrderBy, searchReq.PageInfo.PageSize)
		if err != nil {
			ctx.SendErr(proto.CodeFailedToOperateDB, nil)
			return
		}
		ctx.SendOk(&rspData)
		return
	} else {
		ctx.SendErr(proto.CodeReqParamErr, nil)
		return
	}

}

// GetFileInfo
// @Summary Query file info
// @Description Query file info
// @Tags File
// @Accept application/json
// @Produce application/json
// @Param userId query int true "user id"
// @Param uuid query string false "file/folder uuid"
// @Param path query string false "path"
// @Param name query string false "name"
// @Success 200 {object} proto.Rsp{results=proto.FileInfoRsp} "file info"
// @Router /space/v1/api/file/info [get]
func GetFileInfo(c *gin.Context) {

	ctx := bpctx.NewCtx(c)

	var req proto.FileInfoReq
	var rsp proto.FileInfoRsp

	ctx.LogI("getFileInfo", &req)
	if err := c.ShouldBind(&req); err != nil {
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}
	if req.Path == "" && req.Name == "" {
		if info, err := dbutils.GetFileInfoWithUid(ctx.GetUserId(), req.Uuid); err != nil {
			ctx.SendErr(proto.CodeFailedToOperateDB, err)
			return
		} else {
			rsp = info.FileInfoPub
			ctx.SendOk(&rsp)
			return
		}
	} else if len(req.Path) > 0 && len(req.Name) > 0 {
		if info, err := dbutils.GetInfoByPath(ctx.GetUserId(), req.Path, req.Name, 0); err != nil {
			ctx.SendErr(proto.CodeFileNotExist, err)
			return
		} else {
			rsp = info.FileInfoPub
			ctx.SendOk(&rsp)
			return
		}
	} else {
		ctx.SendErr(proto.CodeReqParamErr, nil)
		return
	}

}

// GetFileInfoForInner Get file info for inner
// @Summary Get file info for inner
// @Description Get file info for inner
// @Tags File
// @Accept application/json
// @Produce application/json
// @Param userId query int true "user id"
// @Param uuid query string false "uuid"
// @Success 200 {object} proto.Rsp{results=proto.FileInfoForInnerRsp} ""
// @Router /space/v1/api/inner/file/info [get]
func GetFileInfoForInner(c *gin.Context) {

	ctx := bpctx.NewCtx(c)

	var req proto.FileInfoForInnerReq
	var rsp proto.FileInfoForInnerRsp

	ctx.LogI("getFileInfoForInner", &req)
	if err := c.ShouldBind(&req); err != nil {
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}

	if info, err := dbutils.GetFileInfoForTrends(ctx.GetUserId(), req.Uuid); err != nil {
		ctx.SendErr(proto.CodeFailedToOperateDB, err)
		return
	} else {
		rsp.FileInfoForTrends = *info
		if !info.IsDir {
			if path, err := stor.GetRelativePath(env.NORMAL_BUCKET, rsp.BETag); err != nil {
				ctx.SendErr(proto.CodeFailedToOperateDB, err)
				return
			} else {
				dir, _ := filepath.Split(path)
				rsp.RelativePath = dir
			}

		}
		ctx.SendOk(&rsp)
		return
	}
}

// CreateVodSymlink Create symbolic link
// @Summary Create symbolic link
// @Description Create symbolic link
// @Tags File
// @Accept application/json
// @Produce application/json
// @Param userId query string true "user id"
// @Param VodSymlinkReq body proto.VodSymlinkReq true "parmas"
// @Success 200 {object} proto.Rsp{results=proto.VodSymlinkRsp} ""
// @Router /space/v1/api/file/vod/symlink [POST]
func CreateVodSymlink(c *gin.Context) {
	ctx := bpctx.NewCtx(c)

	var req proto.VodSymlinkReq
	var rsp proto.VodSymlinkRsp

	userId := ctx.GetUserId()

	//获取请求参数
	if err := c.ShouldBind(&req); err != nil {
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}
	defer ctx.LogI("CreateSoftlink", req)

	err := validate.Struct(req)
	if err != nil {
		ctx.LogE().Err(err).Msg("validate failed")
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}

	info, err := dbutils.GetFileInfoWithUid(userId, req.Id)
	if err != nil {
		ctx.SendErr(proto.CodeFailedToOperateDB, err)
		return
	}

	if len(info.BETag) == 0 {
		ctx.SendErr(proto.CodeReqParamErr, fmt.Errorf("%v is folder", req.Id))
		return
	}
	rpath, _ := stor.GetRelativePath(env.NORMAL_BUCKET, info.BETag)
	oldName := filepath.Join("..", rpath)

	dir := filepath.Join(env.DATA_PATH, "symlink")
	os.Mkdir(dir, os.ModePerm)

	linkName := req.Id + fmt.Sprintf("-%v", time.Now().Format("20060102030405"))

	linkPath := filepath.Join(dir, linkName)
	ctx.Append("oldName", oldName)
	if err := os.Symlink(oldName, linkPath); err != nil {
		ctx.SendErr(proto.CodeFailedToCreateSymlink, err)
		return
	}

	rsp.Linkname = linkName
	ctx.SendOk(rsp)
}

// GetFileInfosForTrends Get fileinfo list for inner
// @Summary Get fileinfo list for inner
// @Description Get fileinfo list for inner
// @Tags File
// @Accept application/json
// @Produce application/json
// @Param userId query int true "user id"
// @Param FileInfoForTrendsReq body proto.FileInfoForTrendsReq false "uuids"
// @Success 200 {object} proto.Rsp{results=proto.FileInfoForTrendsRsp} ""
// @Router /space/v1/api/inner/file/infos [POST]
func GetFileInfosForTrends(c *gin.Context) {

	ctx := bpctx.NewCtx(c)

	var req proto.FileInfoForTrendsReq
	var rsp proto.FileInfoForTrendsRsp

	//ctx.LogI("getFileInfoForInner", &req)
	if err := c.ShouldBind(&req); err != nil {
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}

	if info, err := dbutils.GetFileInfosForTrends(ctx.GetUserId(), req.Uuids); err != nil {
		ctx.SendErr(proto.CodeFailedToOperateDB, err)
		return
	} else {
		rsp.FileInfos = info
		ctx.SendOk(&rsp)
		return
	}
}
