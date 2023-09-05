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

package proto

//此文件定义分片上传

//创建分片任务
type CreateMultipartTaskParam struct {
	FileName   string `json:"fileName" form:"fileName" validate:"required,gt=0"`
	Size       int64  `json:"size" form:"size" validate:"required,gt=0"`
	FolderId   string `json:"folderId" form:"folderId" `
	FolderPath string `json:"folderPath" form:"folderPath" `
	BETag      string `json:"betag" form:"betag" validate:"required,gte=32,lte=34"`
	Mime       string `json:"mime" form:"mime"`
	CreateTime int64  `json:"createTime" form:"modifyTime"`
	ModifyTime int64  `json:"modifyTime" form:"modifyTime"`
	BusinessId int    `json:"businessId" form:"businessId"` 
	AlbumId    int    `json:"albumId" form:"albumId"`
}

type CreateMultipartTaskReq = CreateMultipartTaskParam

type CreateMultipartTaskSuccRsp struct {
	UploadId string `json:"uploadId" form:"uploadId"`
	PartSize int64  `json:"partSize"`
}

type CreateMultipartTaskConflictRsp struct {
	CreateMultipartTaskParam
	ListMultipartRsp
	UploadId string `json:"uploadId" form:"uploadId"`
}

type MultipartTaskStatusInfo = CreateMultipartTaskConflictRsp

const (
	CREATE_MULTIPART_TASK_SUCC     int = 0
	CREATE_MULTIPART_TASK_COMPLETE int = 1
	CREATE_MULTIPART_TASK_CONFLICT int = 2
)

type CreateMultipartTaskRsp struct {
	RspType      int                             `json:"rspType" form:"rspType"`           //rsp type
	SuccInfo     *CreateMultipartTaskSuccRsp     `json:"succInfo" form:"succInfo"`         //task info
	CompleteInfo *FileInfo                       `json:"completeInfo" form:"completeInfo"` //task completed
	ConflictInfo *CreateMultipartTaskConflictRsp `json:"conflictInfo" form:"conflictInfo"` //task exists
}

type MultipartTaskId struct {
	UploadId string `json:"uploadId" form:"uploadId" validate:"required,len=34"`
}

//删除分片任务
type DeleteMultipartTaskReq = MultipartTaskId

//无响应体

//查询已上传分片信息
type ListMultipartReq = MultipartTaskId
type Part struct {
	Start int64 `json:"start" form:"start" validate:"gte=0"`
	End   int64 `json:"end" form:"end" validate:"required,gt=0"`
}

func (part *Part) Len() int64 {
	return part.End - part.Start + 1
}

type ListMultipartRsp struct {
	UploadedParts  []Part `json:"uploadedParts" form:"uploadedParts"`
	UploadingParts []Part `json:"uploadingParts" form:"uploadingParts"`
}

//上传分片
type UploadPartReq struct {
	MultipartTaskId
	Part
	Md5sum string `json:"md5sum" form:"md5sum" validate:"required,len=32"`
}

//回应无包体

//上传完成合并分片
type CompleteMultipartTaskReq = MultipartTaskId
type CompleteMultipartTaskRsp = FileInfo
