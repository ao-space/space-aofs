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
	"aofs/internal/proto"
	"errors"
	"fmt"
	"strings"
	"time"
)

// MoveFiles 移动文件（夹）
func MoveFiles(userId proto.UserIdType, moveId string, destPathId string) (affect int, err error) {
	tx := db.Begin()
	var fileOrDir proto.FileInfo
	//var dstPath proto.FileInfo
	var oriPath proto.FileInfo
	var destPath string
	var subUuids []string
	tx.Model(&fileOrDir).Where("user_id = ? AND uuid = ? AND trashed = ?", userId, moveId, 0).First(&fileOrDir)
	if !fileOrDir.IsDir {

		// 获取目的路径path
		destPath, err = GetAbsPath(userId, destPathId)
		if err != nil {
			return 0, err
		}
		res := tx.Model(&proto.FileInfo{}).Where("user_id = ? AND uuid =? AND trashed = ?", userId, moveId, 0).Updates(map[string]interface{}{"parent_uuid": destPathId, "path": destPath, "operation_time": time.Now().UnixNano() / 1e6})

		affect = int(res.RowsAffected)
		err = res.Error
	} else {
		// 获取目的路径path
		destPath, err = GetAbsPath(userId, destPathId)
		if err != nil {
			return 0, err
		}

		destPathLen := len(strings.Split(destPath, "/"))

		ownChildFolderMaxLen, _ := GetSubDirMaxLayer(userId, moveId)

		if destPathLen+ownChildFolderMaxLen >= 20 {
			return 0, errors.New("beyond 20 layers")
		}

		// 获取原始路径path
		tx.Model(&proto.FileInfo{}).Where("user_id = ? AND uuid = ? AND trashed = ?", userId, moveId, 0).First(&oriPath)
		originPath := oriPath.Path + oriPath.Name + "/"

		err = tx.Model(&proto.FileInfo{}).Where("uuid = ? AND trashed = ?", moveId, 0).Updates(map[string]interface{}{"parent_uuid": destPathId, "path": destPath, "operation_time": time.Now().UnixNano() / 1e6}).Error
		if err != nil {
			return 0, err
		}
		// 拼接sql
		sql := fmt.Sprintf("UPDATE \"aofs_file_infos\" SET path = replace(path,'%s','%s') where user_id = '%d' AND path like '%s'", originPath, destPath+oriPath.Name+"/", userId, originPath+"%")
		tx.Model(&proto.FileInfo{}).Select("uuid").Where("user_id = ? AND path LIKE ? AND category IN (?,?)", userId, oriPath.Path+oriPath.Name+"/%", "picture", "video").Find(&subUuids)

		// 更新该目录下所有文件的path字段
		res := tx.Exec(sql)

		affect = int(res.RowsAffected + 1)
		err = res.Error
	}

	if err != nil {
		tx.Rollback()
		return 0, err
	} else {
		tx.Commit()
		return affect, nil
	}

}

// GetChildFolderMaxLayer 获取子文件夹层数
func GetSubDirMaxLayer(userId proto.UserIdType, uuid string) (int, error) {
	fi, err := GetInfoByUuid(uuid)
	if err != nil {
		return 0, err
	}
	if !fi.IsDir {
		return 0, err
	}
	var subDirs []string
	db.Model(&proto.FileInfo{}).Where("path LIKE ? AND is_idr = ?", fi.AbsPath()+"%", true).Select("uuid").
		Find(&subDirs)
	maxLayer := 0
	for _, subDir := range subDirs {
		subFi, err := GetInfoByUuid(subDir)
		if err != nil {
			continue
		}
		absPath := subFi.AbsPath()
		layer := len(strings.Split(absPath, "/"))
		if layer > maxLayer {
			maxLayer = layer
		}
	}
	return maxLayer, nil
}
