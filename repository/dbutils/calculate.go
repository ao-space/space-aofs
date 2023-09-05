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
	"path/filepath"
	"strings"
)

func CalculateFolderSize(path string, name string) (size int64, err error) {
	// 更新目录的size，可在每次上传完成后执行
	pathWithoutSlash := strings.TrimRight(path, "/")
	parentFolder, folderName := filepath.Split(pathWithoutSlash)
	// 当前目录的计算
	sql := fmt.Sprintf("UPDATE \"aofs_file_infos\" SET size=(SELECT sum(size) FROM \"aofs_file_infos\" WHERE path LIKE '%s'||'/%%' AND trashed IN (0,1)) WHERE path='%s' AND name='%s' AND trashed=0", pathWithoutSlash, parentFolder, folderName)
	res := db.Exec(sql)
	affect := int(res.RowsAffected)
	if affect == 1 {
		var subQuery proto.FileInfo
		db.Model(&subQuery).Where("path = ? AND name = ? AND trashed = 0", parentFolder, folderName).First(&subQuery)
		size = subQuery.Size
		return size, nil
	} else {
		return 0, err
	}
}

func CalculateFolderSizeWithUuid(userId proto.UserIdType, folderUuid string) (size int64, err error) {
	pathRecord, nameRecord := GetPathByUuid(folderUuid)
	sql := fmt.Sprintf("UPDATE \"aofs_file_infos\" SET size=(SELECT sum(size) FROM \"aofs_file_infos\" WHERE path LIKE '%s/%%' AND trashed = 0)  WHERE path='%s' AND name='%s' AND trashed=0 AND user_id='%d' ", pathRecord+nameRecord, pathRecord, nameRecord, userId)
	res := db.Exec(sql)
	affect := int(res.RowsAffected)
	if affect == 1 {
		var subQuery proto.FileInfo
		err = db.Model(&subQuery).Where("path = ? AND name = ? AND trashed = 0 AND user_id = ?", pathRecord, nameRecord, userId).First(&subQuery).Error
		size = subQuery.Size
		return size, nil
	} else {
		return 0, err
	}

}

// CalculateFileCount 计算文件夹内的文件数
func CalculateFileCount(userId proto.UserIdType, path string, name string) (int64, error) {
	var fileCount int64
	//db.Model(&proto.FileInfo{}).Where("path = ?", path).Where("trashed = ?", 0).Count(&fileCount)
	db.Model(&proto.FileInfo{}).Where("path = ? AND trashed = ? AND user_id = ?", path, 0, userId).Count(&fileCount)
	pathWithoutSlash := strings.TrimRight(path, "/")
	parentFolder, folderName := filepath.Split(pathWithoutSlash)
	//res := db.Model(&proto.FileInfo{}).Where("path = ?", parentFolder).Where("name = ?", folderName).Where("trashed = ?", 0).Update("file_count", fileCount)
	res := db.Model(&proto.FileInfo{}).Where("path = ? AND name = ? AND trashed = ? AND user_id = ?", parentFolder, folderName, 0, userId).Update("file_count", fileCount)
	err := res.Error
	if err != nil {
		return 0, err
	}
	return fileCount, nil
}

func CountFileInFolder(uuid string) (fileCount int64) {
	//var fileInfo []proto.FileInfo
	//db.Model(&proto.FileInfo{}).Where("parent_uuid = ? AND trashed = ?", uuid, 0).Scan(&fileInfo)
	db.Model(&proto.FileInfo{}).Where("parent_uuid = ? AND trashed = ?", uuid, 0).Count(&fileCount)
	return fileCount
}

