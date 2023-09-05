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

// src/lib/proto/folder.go
// GetListRspBody 获取文件列表返回值
/*
此文件定义文件夹类消息协议
*/

//CreateFolderReq --- 创建文件夹请求参数
type CreateFolderReq struct {
	FolderName     string `json:"folderName" form:"folderName"`         // folder name
	CurrentDirUuid string `json:"currentDirUuid" form:"currentDirUuid"` // current folder's uuid
}

//DeleteFolderReq --- 删除文件夹请求参数
type DeleteFolderReq struct {
	DeleteId []string `json:"uuid" form:"uuid"` // uuid
}

// RenameFolderReq --- 重命名文件夹请求参数
type RenameFolderReq struct {
	NewName    string `json:"folderName" form:"folderName"` // 新文件夹名
	FolderUuid string `json:"uuid" form:"uuid"`             // 选择的文件夹的UUID
}

// MoveFolderReq --- 移动文件夹请求参数
type MoveFolderReq struct {
	DestPath   string   `json:"destPath" form:"destPath"` // 移动的目标路径
	FolderUuid []string `json:"uuid" form:"uuid"`         // 选择的文件夹的UUID列表
}

type CopyFolderReq struct {
	DestPath   string   `json:"destPath" form:"destPath"` // 复制的目标路径
	FolderUuid []string `json:"uuid" form:"uuid"`         // 选择的文件夹的UUID列表
}

type FolderInfo struct {
	FolderName string `gorm:"column:name" json:"name" form:"name"`
	Path       string `gorm:"column:path" json:"path" form:"path"`
	FolderSize int64  `gorm:"column:size" json:"size" form:"size"`
	UpdateTime int64  `gorm:"column:operation_time" json:"operationAt" form:"operationAt"`
}

type FolderInfoReq struct {
	FolderUuid string `json:"uuid" form:"uuid"`
}
