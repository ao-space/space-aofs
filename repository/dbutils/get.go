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
)

func GetFileInfoWithUid(userId proto.UserIdType, uuid string) (*proto.FileInfo, error) {
	var file proto.FileInfo
	value := db.First(&file, "user_id=? AND uuid = ?", userId, uuid)
	if value.Error != nil {
		return nil, value.Error
	}
	return &file, nil
}

func GetFileInfoForTrends(userId proto.UserIdType, uuid string) (*proto.FileInfoForTrends, error) {
	var file proto.FileInfoForTrends
	value := db.First(&file, "user_id= ? AND uuid = ?", userId, uuid)
	db.Debug().Model(&proto.FileInfo{}).
		Select("\"aofs_file_infos\".*, e.duration").
		Joins("left join t_photo_exif e on \"aofs_file_infos\".uuid = e.uuid ").
		Where("\"aofs_file_infos\".user_id = ? AND \"aofs_file_infos\".uuid = ?", userId, uuid).
		Scan(&file)

	if value.Error != nil {
		return nil, value.Error
	}
	return &file, nil
}

func GetFileInfosForTrends(userId proto.UserIdType, uuids []string) ([]proto.FileInfoForTrends, error) {
	var file []proto.FileInfoForTrends
	//value := db.First(&file, "user_id= ? AND uuid = ?", userId, uuid)
	res := db.Model(&proto.FileInfo{}).
		Select("\"aofs_file_infos\".*, e.duration").
		Joins("left join t_photo_exif e on \"aofs_file_infos\".uuid = e.uuid ").
		Where("\"aofs_file_infos\".user_id = ? AND \"aofs_file_infos\".uuid IN (?)", userId, uuids).
		Scan(&file)

	if res.Error != nil {
		return nil, res.Error
	}
	return file, nil
}

func GetAbsPath(userId proto.UserIdType, folderId string) (string, error) {
	var folderInfo proto.FileInfo
	if len(folderId) != 0 {
		fi := db.Model(&proto.FileInfo{}).Where("user_id = ? AND uuid = ? AND trashed = ?", userId, folderId, 0).First(&folderInfo)
		if fi.Error != nil {
			return "", fi.Error
		}
	} else {
		return "/", nil
	}
	return folderInfo.AbsPath(), nil
}


func GetFilesInUuids(userId proto.UserIdType, uuids []string) (files []string, err error) {
	condition := map[string]interface{}{}
	condition["uuid"] = uuids
	condition["is_dir"] = false
	condition["user_id"] = userId
	err = db.Model(&proto.FileInfo{}).Where(condition).Select("uuid").Scan(&files).Error
	if err != nil {
		return nil, err
	}
	return files, nil
}

func GetSubFilesInUuids(userId proto.UserIdType, uuids []string, trashed []uint32) (allSubFiles []string, err error) {
	var dirs []string
	var subFiles []string
	condition := map[string]interface{}{}
	condition["uuid"] = uuids
	condition["is_dir"] = true
	condition["user_id"] = userId
	// 找到 dir 的 uuids
	err = db.Model(&proto.FileInfo{}).Where(condition).Select("uuid").Scan(&dirs).Error
	if err != nil {
		return nil, err
	}
	if len(dirs) == 0 {
		return nil, nil
	}
	// 根据dir uuid 找到所有目录下 文件的uuid
	for _, dir := range dirs {
		fi, err := GetInfoByUuid(dir)
		if err == nil {
			path := fi.AbsPath()
			err = db.Debug().Model(&proto.FileInfo{}).Where("user_id = ? AND path LIKE ? AND trashed IN (?)", userId, path+"%", trashed).Select("uuid").Scan(&subFiles).Error
			if err != nil {
				return nil, err
			}
			allSubFiles = append(allSubFiles, subFiles...)
		}
	}
	logdb.LogD().Interface("subUuids", allSubFiles).Msg("print sub  uuids")
	return allSubFiles, nil
}
