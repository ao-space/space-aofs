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

/*
此文件定义文件类消息协议
*/

//

/*
--- MoveFileReq 移动文件协议

{
    "id": "91c8c5c3-ad46-4721-9687-09df8f2e88c2", # 文件的id
    "dest-path": "/picture"
}
*/

type MoveFileReq struct {
	Id       []string `json:"uuids" form:"uuids" validate:"required,unique"` // uuid
	DestPath string   `json:"destPath" form:"destPath" `                     // dest path
}

// type MoveFileRspBody = FileInfo

type GetListReq struct {
	PageInfo
	Uuid     string `json:"uuid" form:"uuid"`
	IsDir    bool   `json:"isDir" form:"isDir"`
	OrderBy  string `json:"orderBy" form:"orderBy"`
	Category string `json:"category" form:"category"`
}

type GetListRspData struct {
	List     []FileInfoPub `json:"fileList" form:"fileList"`
	PageInfo PageInfoExt   `json:"pageInfo" form:"pageInfo"`
}

// DeleteFileReq 删除文件请求参数 文件uuid
type DeleteFileReq struct {
	DeleteIds []string `json:"uuids" form:"uuids" validate:"unique"` // uuids
}

// CopyFileReq 复制文件
type CopyFileReq struct {
	Ids    []string `json:"uuids" form:"uuids" validate:"unique,gt=0"` // uuids
	DestId string   `json:"dstPath" form:"dstPath"`                    // dest path
}

// ModifyFileReq 修改文件名
type ModifyFileReq struct {
	Id          string `json:"uuid" form:"uuid" validate:"uuid"` // file/folder uuid
	NewFileName string `json:"fileName" form:"fileName"`         // new name
}

// DbAffect 数据库操作受影响的行数
type DbAffect struct {
	AffectRows uint32 `json:"affectRows" form:"affectRows"`
}

type CopyRsp struct {
	AffectRows uint32 `json:"affectRows" form:"affectRows"`
	Data       []NewAndOldUuid
}

// NewAndOldUuid 复制操作会产生新的记录，新的uuid
type NewAndOldUuid struct {
	OldId string `gorm:"column:uuid" json:"oldId" form:"oldId"`
	NewId string `gorm:"column:uuid" json:"newId" form:"newId"`
}

// PathName --查询返回值
type PathName struct {
	Path string `json:"path" form:"path"`
	Name string `json:"name" form:"name"`
}

// Uuids --查询返回值，复制文件夹时候用到
type Uuids struct {
	CopyUuid string `gorm:"column:uuid" json:"copyUuid" form:"copyUuid"`
}

type SearchReq struct {
	PageInfo
	Uuid       string `json:"uuid" form:"uuid"`         //当前文件夹的uuid
	ObjectName string `json:"name" form:"name"`         //搜索的对象名
	Category   string `json:"category" form:"category"` //分类
	OrderBy    string `json:"orderBy" form:"orderBy"`   //排序
}

//获取单个文件信息
type FileMetaReq = PathName
type FileInfoReq struct {
	Fid
	PathName
}
type FileInfoRsp = FileInfoPub
type FileMetaRsp = FileInfo
type FileInfoForInnerReq = Fid
type FileInfoForTrendsReq = Fids
type FileInfoForInnerRsp struct {
	FileInfoForTrends
	RelativePath string `json:"relativePath" form:"relativePath"`
}

type FileInfoForTrendsRsp struct {
	FileInfos []FileInfoForTrends `json:"fileInfos" form:"fileInfos"`
}

// ModifyFileReq 修改文件名
type VodSymlinkReq struct {
	Id string `json:"uuid" form:"uuid" validate:"uuid"` //  file's uuid
}

type VodSymlinkRsp struct {
	Linkname string `json:"linkName" form:"linkName"`
}

type FileInfoForTrends struct {
	FileInfo
	Duration int64 `json:"duration" form:"duration"`
}

type FileChangePushMsg struct {
	OperatorType string   `json:"operator_type,omitempty"`
	Uuids        []string `json:"uuids,omitempty"`
}
