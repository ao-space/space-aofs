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
	"fmt"
	"time"
)

// RenameFiles  修改文件 ,文件夹名，合并
func RenameFiles(userId proto.UserIdType, uuid string, name string) (affect int64, err error) {

	tx := db.Begin()
	var FileOrDir proto.FileInfo
	tx.Model(&FileOrDir).Where("user_id = ? AND uuid = ? AND trashed = ?", userId, uuid, 0).First(&FileOrDir)
	if !FileOrDir.IsDir {
		res := tx.Model(&proto.FileInfo{}).Where("user_id = ? AND uuid = ? AND trashed = ?", userId, uuid, 0).Updates(map[string]interface{}{"name": name, "operation_time": time.Now().UnixNano() / 1e6})
		affect = res.RowsAffected
		err = res.Error
	} else {
		// 获取原始路径path
		var parentPath proto.FileInfo
		tx.Model(&proto.FileInfo{}).Where("user_id = ? AND uuid = ? AND trashed = ?", userId, uuid, 0).First(&parentPath)
		originPath := parentPath.Path + parentPath.Name + "/"
		if len(originPath) != 0 {
			// 更新文件夹名name
			res := tx.Model(&proto.FileInfo{}).Where("user_id = ? AND uuid = ? AND trashed = ?", userId, uuid, 0).Updates(map[string]interface{}{"name": name, "operation_time": time.Now().UnixNano() / 1e6})
			if res.Error != nil {
				return 0, err
			} else {
				affect += 1
			}
		}
		// 拼接sql
		sql := fmt.Sprintf("UPDATE \"aofs_file_infos\" SET path = replace(path,'%s','%s') where user_id=%d AND trashed = 0 AND path like '%s'", originPath, parentPath.Path+name+"/", userId, originPath+"%")
		fmt.Println(sql)
		//更新该文件夹下所有文件的path
		isSubFileExist := tx.Model(&proto.FileInfo{}).Where("user_id = ? AND path = ? AND trashed = ?", userId, originPath, 0).First(&proto.FileInfo{})


		if isSubFileExist.RowsAffected != 0 {
			res := tx.Exec(sql)
			affect = res.RowsAffected + 1
			err = res.Error
		}
	}

	if err != nil {
		tx.Rollback()
		return 0, err
	} else {
		tx.Commit()
		return affect, nil
	}

}


func UpdateOperationTime(uuid string) error {
	err := db.Model(&proto.FileInfo{}).Where("uuid = ?", uuid).Update("operation_time", time.Now().UnixNano()/1e6).Error
	return err
}


func UpdateFileInfoExt(charset []byte, uuid string) error {
	err := db.Model(&proto.FileInfo{}).Where("uuid = ? ", uuid).Update("ext", charset).Error
	if err != nil {
		return err
	}
	return nil
}
