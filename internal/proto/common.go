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

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
)

// FileInfo --- 文件信息结果定义
/*
{
    "id": "6665db67-8c77-4944-af81-f5d90a01c52f", # 文件、目录 的 id
    "type": "file", # 文件类型: file, directory
    "name": "greeting.txt", # 文件名称
    "path": "/Documents/phone", # 文件路径，根路径为 '/'
	"user_id": "用户id，区分空间所属"
    "trashed": false, # 是否为已回收文件或目录
    "md5sum": "5d41402abc4b2a76b9719d911017c592", # 文件内容 md5 摘要
    "created_at": "2021-06-10T01:00:00Z", # 文件的创建时间
    "updated_at": "2021-06-10T01:00:00Z", # 文件的最后修改时间
    "tags": ["greet"], # 文件的 tag （保留属性）
    "size": 200, # 文件的大小
    "executable": false, # 是否为可执行文件
    "Category": "document", # 文件分类，document, picture, audio, video
    "mime": "text/plain", # 文件的类型，mime 格式
    "version": 1 # 文件的当前版本，每次修改都需要提升该版本号
}
*/

type Fid struct {
	Uuid string `json:"uuid" form:"uuid"` // uuid
}
type Fids struct {
	Uuids []string `json:"uuids" form:"uuids"` // uuids
}

type UserIdType uint

type FileInfoPub struct {
	Id            string `gorm:"column:uuid;PRIMARY_KEY;index" json:"uuid" form:"uuid"`
	ParentUuid    string `gorm:"column:parent_uuid" json:"parentUuid" form:"parentUuid" `
	IsDir         bool   `gorm:"column:is_dir" json:"isDir" form:"isDir" `
	Name          string `gorm:"column:name;uniqueIndex:totalPath" json:"name" form:"name"`
	Path          string `gorm:"column:path;uniqueIndex:totalPath" json:"path" form:"path"`
	BETag         string `gorm:"column:betag" json:"betag" form:"betag"`
	CreateTime    int64  `gorm:"column:created_time" json:"createdAt" form:"createdAt"`
	ModifyTime    int64  `gorm:"column:modify_time" json:"modifyAt" form:"modifyAt"`
	OperationTime int64  `gorm:"column:operation_time" json:"operationAt" form:"operationAt"`
	Size          int64  `gorm:"column:size" json:"size" form:"size"`
	Category      string `gorm:"column:category" json:"category" form:"category"`
	Mime          string `gorm:"column:mime" json:"mime" form:"mime"`
	Trashed       uint32 `gorm:"column:trashed;uniqueIndex:totalPath" json:"trashed" form:"trashed"` //0-normal; 1-Logical delete, put into the recycle bin; 2-Has been cleared from the recycle bin and is to be physically deleted
	FileCount     uint32 `gorm:"column:file_count" json:"fileCount" form:"fileCount"`
}

type FileInfoPubLst []FileInfoPub

type FileInfo struct {
	FileInfoPub

	UserId        UserIdType     `gorm:"column:user_id;uniqueIndex:totalPath" json:"userId" form:"userId"`
	Tags          string         `gorm:"column:tags" json:"tags" form:"tags"`
	Executable    bool           `gorm:"column:executable" json:"executable" form:"executable"`
	Version       uint32         `gorm:"column:version" json:"version" form:"version"`
	BucketName    string         `gorm:"column:bucketname" json:"bucketName" form:"bucketName"`
	TransactionId int64          `gorm:"column:transaction_id;default:0" json:"transactionId" form:"transactionId"`
	FileInfoExt   datatypes.JSON `gorm:"column:ext" json:"ext" form:"ext"`
}

type FileInfoExt struct {
	Charset string `json:"charset" form:"charset"`
}

type FileInfoLst []FileInfo

func (lst FileInfoLst) ToPubLst() FileInfoPubLst {
	var lstPub FileInfoPubLst
	for _, fi := range lst {
		lstPub = append(lstPub, fi.FileInfoPub)
	}
	return lstPub
}

func (FileInfo) TableName() string {
	return "aofs_file_infos"
}

func (fi *FileInfo) FixTime() {
	microSeconds := time.Now().UnixNano() / 1e6

	if fi.CreateTime == 0 {
		fi.CreateTime = microSeconds
	}
	if fi.ModifyTime == 0 {
		fi.ModifyTime = microSeconds
	}
	if fi.OperationTime == 0 {
		fi.OperationTime = microSeconds
	}
}

func (fi *FileInfo) AbsPath() string {
	if fi.IsDir {
		if len(fi.Path) == 0 {
			return fi.Name
		} else {
			return fi.Path + fi.Name + "/"
		}
	} else {
		return fi.Path + fi.Name
	}
}

func (fi *FileInfo) EncodeByteFileInfo() ([]byte, error) {

	if byteFileInfo, err := json.Marshal(fi); err != nil {
		return nil, err
	} else {
		return byteFileInfo, nil
	}

}

func DecodeFileInfo(data []byte) (fileInfos []FileInfo, err error) {

	if err = json.Unmarshal(data, &fileInfos); err != nil {
		return nil, err
	}
	return fileInfos, nil
}



type SyncInfo struct {
	DeviceId   string     `gorm:"column:device_id;uniqueIndex:user_device" json:"deviceId" form:"deviceId"`
	DeviceName string     `gorm:"column:device_name" json:"deviceName" form:"deviceName"`
	FolderId   string     `gorm:"column:folder_id" json:"folderId" form:"folderId"`
	UserId     UserIdType `gorm:"column:user_id;uniqueIndex:user_device" json:"userId" form:"userId"`
}

func (SyncInfo) TableName() string {
	return "aofs_sync_infos"
}

const (
	TrashStatusNormal               uint32 = iota //正常文件状态
	TrashStatusLogicDeleted                       //逻辑删除放入回收站
	TrashStatusPhyDeleted                         //从回收站删除/清空回收站, 待物理清除
	TrashStatusPhyDelException                    //从回收站删除/清空回收站, 物理清除异常
	TrashStatusSubFilesLogicDeleted               //逻辑删除的文件夹下的子文件或文件夹
)

