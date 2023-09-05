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

package dbutils

import (
	"errors"
	"aofs/internal/proto"
	"aofs/internal/utils"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// AddFile 上传文件
func AddFile(info proto.FileInfo) (err error) {
	res := db.Model(&proto.FileInfo{}).Create(&info)
	err = res.Error
	if err != nil {

		return err
	}
	return nil
}

// AddFileV2 上传文件,path 传入参数为当前目录的uuid
func AddFileV2(info proto.FileInfo, uuid string) (err error) {
	fi, err := GetInfoByUuid(uuid)
	if err != nil {
		return err
	}

	info.Path = fi.AbsPath()
	return AddFile(info)
}


// CopyFile 复制文件
func CopyFile(userId proto.UserIdType, req proto.CopyFileReq) (affect int, uuids []proto.NewAndOldUuid, err error) {

	var subQuery proto.FileInfo
	var destPath string

	for _, fileId := range req.Ids {
		//idCount := len(req.Id)

		var fileOrDir proto.FileInfo
		exist := db.Model(&fileOrDir).Where("user_id = ? AND uuid = ? AND trashed = 0", userId, fileId).First(&fileOrDir)
		err := exist.Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, uuids, err
		} else {
			// 复制文件
			if !fileOrDir.IsDir {
				// 获取目的路径path
				// 判断是否根目录
				if req.DestId != "" {
					db.Model(&proto.FileInfo{}).Where("user_id = ? AND uuid = ? AND trashed = 0", userId, req.DestId).First(&subQuery)
					destPath = subQuery.Path + subQuery.Name + "/"
				} else {
					destPath = "/"
				}

				var srcInfo proto.FileInfo
				db.Model(&proto.FileInfo{}).Where("user_id = ? AND uuid = ? AND trashed = 0", userId, fileId).Scan(&srcInfo)

				copyInfo := srcInfo
				copyInfo.Id = utils.RandomID()
				copyInfo.Path = destPath
				copyInfo.OperationTime = time.Now().UnixNano() / 1e6
				res := db.Model(&proto.FileInfo{}).Create(&copyInfo)
				//db.Model(&proto.FileInfo{}).Where("uuid = ?", copyInfo.Id).Update("path", destPath)
				affect += int(res.RowsAffected)
				err = res.Error

				uid := proto.NewAndOldUuid{
					OldId: fileId,
					NewId: copyInfo.Id}

				uuids = append(uuids, uid)

			} else {
				// 获取目的路径path
				// 判断是否根目录
				if req.DestId != "" {
					db.Model(&proto.FileInfo{}).Where("user_id = ? AND uuid = ? AND trashed = 0", userId, req.DestId).First(&subQuery)
					destPath = subQuery.Path + subQuery.Name + "/"
				} else {
					destPath = "/"
				}

				destPathLen := len(strings.Split(destPath, "/"))

				ownChildFolderMaxLen, _ := GetSubDirMaxLayer(userId, fileId)

				if destPathLen+ownChildFolderMaxLen >= 20 {
					return 0, nil, errors.New("beyond 20 layers")
				}

				var srcInfo proto.FileInfo
				// 当前目录的path修改
				db.Model(&proto.FileInfo{}).Where("user_id = ? AND uuid = ? AND trashed = 0", userId, fileId).First(&srcInfo)
				copyInfo := srcInfo
				copyInfo.Id = utils.RandomID()
				copyInfo.Path = destPath
				copyInfo.OperationTime = time.Now().UnixNano() / 1e6
				//copyInfo.ParentId = req.DestPath
				if res := db.Model(&proto.FileInfo{}).Create(&copyInfo); res.RowsAffected == 0 {
					return 0, nil, res.Error
				}
				//affect = 1
				// 子目录和文件复制
				originPath := srcInfo.Path + srcInfo.Name + "/"
				var copyId []string
				//var fileList
				db.Model(&proto.FileInfo{}).Select("uuid").Where("user_id = ? AND path LIKE ?", userId, originPath+"%").Find(&copyId)
				fileReq := proto.CopyFileReq{Ids: copyId, DestId: copyInfo.Id}
				affect, _, _ = CopyFile(userId, fileReq)
				uid := proto.NewAndOldUuid{
					OldId: fileId,
					NewId: copyInfo.Id}

				uuids = append(uuids, uid)

			}
		}
		affect = affect + 1

	}
	if err != nil {
		return 0, nil, err
	} else {
		return affect, uuids, nil
	}

}

func RecursiveCreateFolder(userId proto.UserIdType, path string) (*proto.FileInfo, error) {
	if len(path) == 0 || path[0] != '/' || path[len(path)-1] != '/' {
		return nil, fmt.Errorf("path %s is invalid", path)
	}

	lst := strings.Split(path, "/")

	parentInfo, err := GetInfoByPath(userId, "", "/", proto.TrashStatusNormal)
	if err != nil {
		return nil, err
	}

	for _, name := range lst {
		name = strings.TrimSpace(name)
		if len(name) == 0 {
			continue
		}

		var pi proto.FileInfo
		tx := db.Model(&pi).Where("user_id = ? AND path = ? and name = ? and trashed = ?", userId, parentInfo.AbsPath(), name, 0).Scan(&pi)
		if tx.Error != nil {
			return nil, tx.Error
		} else if tx.RowsAffected == 0 {
			//目录不存在，创建之
			if fi, err := CreateFolderByPath(userId, parentInfo.Id, name); err != nil {
				return nil, err
			} else {
				parentInfo = fi
			}

		} else if pi.IsDir {
			parentInfo = &pi
		} else {
			return nil, fmt.Errorf("%s/%s is not dir", pi.Path, pi.Name)
		}
	}
	return parentInfo, nil
}

func CreateFolderByPath(userId proto.UserIdType, uuidParent string, name string) (*proto.FileInfo, error) {
	fi := &proto.FileInfo{}
	tx := db.Model(fi).Where("user_id = ? AND uuid = ?", userId, uuidParent).First(fi)
	if tx.Error != nil {
		return nil, tx.Error
	} else if tx.RowsAffected == 0 {
		return nil, fmt.Errorf("uuid(%s) is not exist", uuidParent)
	} else if !fi.IsDir {
		return nil, fmt.Errorf("uuid(%s) is not dir.", uuidParent)
	}

	currentPath := fi.Path + fi.Name
	if fi.Name != "/" {
		currentPath += "/"
	}

	fi = &proto.FileInfo{
		FileInfoPub: proto.FileInfoPub{
			Id:            utils.RandomID(),
			ParentUuid:    uuidParent,
			IsDir:         true,
			Name:          name,
			Path:          currentPath,
			BETag:         "",
			CreateTime:    time.Now().UnixNano() / 1e6,
			ModifyTime:    time.Now().UnixNano() / 1e6,
			OperationTime: time.Now().UnixNano() / 1e6,
			Size:          0,
			Category:      "",
			Mime:          "",
		},
		UserId:     userId,
		Tags:       "",
		Executable: false,
		Version:    1,
		BucketName: "eulixspace-files",
	}
	tx = db.Model(&proto.FileInfo{}).Create(fi)
	if tx.Error != nil {
		return nil, tx.Error
	} else if tx.RowsAffected > 0 {
		return fi, nil
	} else {
		return nil, fmt.Errorf("unknown error")
	}

}

// CreateFolder 创建文件夹
func CreateFolder(userId proto.UserIdType, req proto.CreateFolderReq) (newFolderInfo proto.FileInfo, returnCode uint8, err error) {
	var subQuery proto.FileInfo
	var currentPath string
	var affect int64
	var rootInfo *proto.FileInfo
	if req.CurrentDirUuid == "" {
		currentPath = "/"
		rootInfo, _ = GetInfoByPath(userId, "", "/", 0)
		req.CurrentDirUuid = rootInfo.Id
	} else {
		db.Model(subQuery).Where("uuid = ?", req.CurrentDirUuid).First(&subQuery)
		currentPath = subQuery.AbsPath() // subQuery.Path + subQuery.Name + "/"
	}

	pathArr := strings.Split(currentPath, "/")
	if len(pathArr) <= 20 {
		err = db.Model(proto.FileInfo{}).Where("user_id = ? AND name = ? AND path = ? AND trashed = 0", userId, req.FolderName, currentPath).First(&proto.FileInfo{}).Error
		if err != nil {
			newFolderInfo = proto.FileInfo{
				FileInfoPub: proto.FileInfoPub{
					Id:            utils.RandomID(),
					ParentUuid:    req.CurrentDirUuid,
					IsDir:         true,
					Name:          req.FolderName,
					Path:          currentPath,
					BETag:         "",
					CreateTime:    time.Now().UnixNano() / 1e6,
					ModifyTime:    time.Now().UnixNano() / 1e6,
					OperationTime: time.Now().UnixNano() / 1e6,
					Size:          0,
					Category:      "",
					Mime:          "",
				},
				UserId:        userId,
				Tags:          "",
				Executable:    false,
				Version:       1,
				BucketName:    "eulixspace-files",
				TransactionId: 0,
			}
			affect = db.Model(&proto.FileInfo{}).Create(&newFolderInfo).RowsAffected
		} else {
			return newFolderInfo, 1, err
		}

	} else {
		return newFolderInfo, 2, errors.New("greater than 20 layers")
	}

	if affect == 1 {
		return newFolderInfo, 3, nil
	}
	return
}

// CopyFolder 复制文件夹
func CopyFolder(req proto.CopyFolderReq) (affect int64, err error) {
	var srcInfo proto.FileInfo
	var subQuery proto.FileInfo
	// 获取目的路径path
	db.Model(&proto.FileInfo{}).Where("uuid = ?", req.DestPath).First(&subQuery)
	destPath := subQuery.Path + subQuery.Name + "/"
	pathArr := strings.Split(destPath, "/")
	if len(pathArr) <= 5 {
		// 循环修改元数据的path字段
		for _, folderId := range req.FolderUuid {
			db.Model(&proto.FileInfo{}).Where("uuid = ?", folderId).First(&srcInfo)
			copyInfo := srcInfo
			copyInfo.Id = utils.RandomID()
			copyInfo.Path = destPath
			res := db.Model(&proto.FileInfo{}).Create(&copyInfo)
			affect = res.RowsAffected
			err = res.Error
		}
	} else {
		return
	}

	if err != nil {
		return 0, err
	}
	return affect, nil
}



