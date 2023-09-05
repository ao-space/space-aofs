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
	"path/filepath"
	"strings"

	"gorm.io/gorm"
)

// GetFileList 获取文件列表
func GetFileList(userId proto.UserIdType, isDir bool, uuid string, page uint32, order string, category string, pageSize int) (fileinfo proto.FileInfoLst, err error) {
	var subQuery []proto.FileInfo
	//获取拼接完整目录路径
	if isDir == false {
		if category == "" {
			db.Model(&fileinfo).Where("uuid = ?", uuid).Scan(&subQuery)
			//fmt.Printf("%#v\\n",path)
			// 获取所有文件列

			fileList := db.Model(&fileinfo).Limit(pageSize).Offset((int(page)-1)*pageSize).
				Where("path = ? AND trashed = ? AND user_id = ?", subQuery[0].Path+subQuery[0].Name+"/", 0, userId).Order(order).Scan(&fileinfo)
			err = fileList.Error
		} else {
			fileList := db.Model(&fileinfo).Limit(pageSize).Offset((int(page)-1)*pageSize).
				Where("category = ? AND trashed = ? AND user_id = ?", category, 0, userId).Order(order).Scan(&fileinfo)
			err = fileList.Error
		}
	} else {
		if category == "" {
			db.Model(&fileinfo).Where("uuid = ?", uuid).Scan(&subQuery)
			//fmt.Printf("%#v\\n",path)
			// 获取所有文件列

			fileList := db.Model(&fileinfo).Limit(pageSize).Offset((int(page)-1)*pageSize).
				Where("path = ? AND trashed = ? AND is_dir = true AND user_id = ?", subQuery[0].Path+subQuery[0].Name+"/", 0, userId).Order(order).Scan(&fileinfo)
			err = fileList.Error
		} else {
			fileList := db.Model(&fileinfo).Limit(pageSize).Offset((int(page)-1)*pageSize).
				Where("category = ? AND trashed = ? AND is_dir = true AND user_id = ?", category, 0, userId).Order(order).Scan(&fileinfo)
			err = fileList.Error
		}
	}

	// println(res)
	if err != nil {
		return nil, err
	}
	return
}

// GetRootList 不提供任何参数时，返回全部文件
func GetRootList(userId proto.UserIdType, isDir bool, page uint32, order string, pageSize uint32) (fileinfo proto.FileInfoLst, err error) {
	//pageSize := 10
	// 获取根目录下所有文件列表
	var fileList *gorm.DB
	if isDir == false {
		fileList = db.Model(&fileinfo).Limit(int(pageSize)).Offset((int(page)-1)*int(pageSize)).Where("path = ? AND trashed = ? AND user_id = ?", "/", 0, userId).Order(order).Scan(&fileinfo)
	} else {
		fileList = db.Model(&fileinfo).Limit(int(pageSize)).Offset((int(page)-1)*int(pageSize)).Where("path = ? AND trashed = ? AND is_dir = true AND user_id = ?", "/", 0, userId).Order(order).Scan(&fileinfo)
	}

	err = fileList.Error
	if err != nil {
		return nil, err
	}
	return
}

// SearchFileByName 根据文件名搜索文件
func SearchFileByName(userId proto.UserIdType, currentDir string, name string, category string, order string, page uint32, pageSize uint32) (fileInfo proto.FileInfoLst, err error) {
	var fileList *gorm.DB
	var categories []string
	if category != "" && strings.Contains(category, ",") {
		categories = strings.Split(category, ",")
	}
	if currentDir == "" && category == "" {
		//fileList = db.Model(&fileInfo).Limit(int(pageSize)).Offset((int(page)-1)*int(pageSize)).Where("name LIKE ? AND trashed = ? ", "%"+name+"%", 0).Order(order).Scan(&fileInfo)
		sql := fmt.Sprintf("SELECT * FROM \"aofs_file_infos\" WHERE name ILIKE '%s' AND trashed = 0 AND user_id = '%d' ORDER BY (CASE\nWHEN name='%s' THEN 1\nWHEN name ILIKE '%s' THEN 2\nWHEN is_dir=true THEN 3\nWHEN name ILIKE '%s' THEN 4\nWHEN name ILIKE '%s' THEN 5\nWHEN name ILIKE '%s' THEN 6\nELSE 7\nEND) limit %d offset (%d-1)*%d ", "%"+"\\"+name+"%", userId, name, name+".%", name+"%", "%"+name+"%", "%"+name, pageSize, page, pageSize)
		//fmt.Println(sql)
		fileList = db.Raw(sql).Scan(&fileInfo)
	} else if category == "" {
		parentPath, folderName := GetPathByUuid(currentDir)
		currentPath := parentPath + folderName + "/"
		//fileList = db.Model(&fileInfo).Limit(int(pageSize)).Offset((int(page)-1)*int(pageSize)).Where("path = ? AND name LIKE ? trashed = ? ", currentPath, "%"+name+"%", 0).Order(order).Scan(&fileInfo)
		sql := fmt.Sprintf("SELECT * FROM \"aofs_file_infos\" WHERE path='%s' AND name ILIKE '%s' AND trashed = 0 AND user_id = '%d' ORDER BY (CASE\nWHEN name='%s' THEN 1\nWHEN name ILIKE '%s' THEN 2\nWHEN is_dir=true THEN 3\nWHEN name ILIKE '%s' THEN 4\nWHEN name ILIKE '%s' THEN 5\nWHEN name ILIKE '%s' THEN 6\nELSE 7\nEND) limit %d offset (%d-1)*%d", currentPath, "%"+"\\"+name+"%", userId, name, name+".%", name+"%", "%"+name+"%", "%"+name, pageSize, page, pageSize)
		fileList = db.Raw(sql).Scan(&fileInfo)
	} else {
		if len(categories) > 1 {
			//fileList = db.Model(&fileInfo).Limit(int(pageSize)).Offset((int(page)-1)*int(pageSize)).Where("name LIKE ? AND trashed = ? AND category = ? ", "%"+name+"%", 0, category).Order(order).Scan(&fileInfo)
			sql := fmt.Sprintf("SELECT * FROM \"aofs_file_infos\" WHERE category IN ('%s','%s') AND name ILIKE '%s' AND trashed = 0 AND user_id = '%d' ORDER BY (CASE\nWHEN name='%s' THEN 1\nWHEN name ILIKE '%s' THEN 2\nWHEN is_dir=true THEN 3\nWHEN name ILIKE '%s' THEN 4\nWHEN name ILIKE '%s' THEN 5\nWHEN name ILIKE '%s' THEN 6\nELSE 7\nEND) limit %d offset (%d-1)*%d", categories[0], categories[1], "%"+"\\"+name+"%", userId, name, name+".%", name+"%", "%"+name+"%", "%"+name, pageSize, page, pageSize)
			fileList = db.Raw(sql).Scan(&fileInfo)
		} else {
			sql := fmt.Sprintf("SELECT * FROM \"aofs_file_infos\" WHERE category = '%s' AND name ILIKE '%s' AND trashed = 0 AND user_id = '%d' ORDER BY (CASE\nWHEN name='%s' THEN 1\nWHEN name ILIKE '%s' THEN 2\nWHEN is_dir=true THEN 3\nWHEN name ILIKE '%s' THEN 4\nWHEN name ILIKE '%s' THEN 5\nWHEN name ILIKE '%s' THEN 6\nELSE 7\nEND) limit %d offset (%d-1)*%d", category, "%"+"\\"+name+"%", userId, name, name+".%", name+"%", "%"+name+"%", "%"+name, pageSize, page, pageSize)
			fileList = db.Raw(sql).Scan(&fileInfo)
		}

	}
	if fileList.Error != nil {
		return nil, fileList.Error
	} else {
		return fileInfo, nil
	}
}

// FileIsExist 判断文件（夹）是否存在
func FileIsExist(userId proto.UserIdType, uuid string, path string, name string) (affect int, err error) {
	var subQuery proto.FileInfo
	if uuid == "" {
		if path != "" || name != "" {
			res := db.Model(subQuery).Where("name = ? AND path = ? AND user_id = ? AND trashed = 0", name, path, userId).First(&subQuery)
			if res.Error != nil {
				return 0, err
			} else {
				return int(res.RowsAffected), nil
			}
		} else {
			res := db.Model(subQuery).Where("name = ? AND path = ? AND user_id = ? AND trashed = 0", "/", "", userId).First(&subQuery)
			if res.Error != nil {
				return 0, err
			} else {
				return int(res.RowsAffected), nil
			}
		}
	} else {
		res := db.Model(subQuery).Where("uuid = ?", uuid).First(&subQuery)
		if res.Error != nil {
			return 0, res.Error
		} else {
			return int(res.RowsAffected), nil
		}
	}
}

// FileIsExistByUuid 根据uuid判断文件是否存在
func FileIsExistByUuid(uuid string) bool {
	var fileInfo proto.FileInfo
	err := db.Model(fileInfo).Where("uuid = ? AND trashed = ?", uuid, 0).First(&fileInfo).Error
	if err != nil {
		return false
	} else {
		return true
	}
}

// IsExistByPath 根据路径判断文件是否存在
func IsExistByPath(userId proto.UserIdType, path string, name string) (bool, error) {
	var subQuery proto.FileInfo
	db := db.Model(subQuery).Where("user_id=? AND name = ? AND path = ? AND trashed IN (0,1)", userId, name, path).Scan(&subQuery)
	if db.Error != nil {
		return false, db.Error
	} else {
		return db.RowsAffected > 0, nil
	}
}

// GetInfoByPath 根据路径获取文件信息
func GetInfoByPath(userId proto.UserIdType, path string, name string, trashed uint32) (*proto.FileInfo, error) {
	var fileInfo proto.FileInfo
	tx := db.Model(fileInfo).Where("user_id = ? AND name = ? AND path = ? AND trashed = ?", userId, name, path, trashed).First(&fileInfo)
	if tx.Error == nil {
		return &fileInfo, nil
	} else { //if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return nil, tx.Error
	}
}

//GenIncNameByPath 根据规则获取新的递增文件名称
func GenIncNameByPath(userId proto.UserIdType, path string, name string, trashed uint32) (string, error) {

	_, err := GetInfoByPath(userId, path, name, trashed)
	if err == nil {
		//按规则处理
		ext := filepath.Ext(name)
		preName := name[:len(name)-len(ext)]
		for i := 1; ; i++ {
			newName := preName + fmt.Sprintf("(%d)", i) + ext
			_, err = GetInfoByPath(userId, path, newName, trashed)
			if err == nil {
				continue
			} else if errors.Is(err, gorm.ErrRecordNotFound) {
				return newName, nil
			} else {
				return "", err
			}
		}

	} else {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return name, nil
		} else {
			return "", nil
		}
	}

}

func GetInfoByUuid(uuid string) (*proto.FileInfo, error) {
	var fileInfo proto.FileInfo
	db := db.Model(fileInfo).Where("uuid = ?", uuid).First(&fileInfo)
	if db.Error != nil {
		return nil, db.Error
	} else if db.RowsAffected == 0 {
		return nil, fmt.Errorf("record not exist")
	} else {
		return &fileInfo, nil
	}
}

func GetPathByUuid(uuid string) (string, string) {
	var pathName proto.PathName
	var fileList proto.FileInfo
	if uuid == "" {
		return "/", ""
	} else {
		db.Model(fileList).Select("path,name").Where("uuid = ?", uuid).First(&pathName)
	}
	//fmt.Printf("%v%v", pathName.Path, pathName.Name)
	return pathName.Path, pathName.Name
}

func GetRecycledList(userId proto.UserIdType, page uint32, order string, pageSize uint32) (fileList proto.FileInfoLst, err error) {
	recycledList := db.Model(&fileList).Limit(int(pageSize)).Offset((int(page)-1)*int(pageSize)).Where("trashed = ? AND user_id = ?", 1, userId).Order("transaction_id DESC").Scan(&fileList)
	err = recycledList.Error
	if err != nil {
		return nil, err
	}
	return
}

func GetRecycledPhyDeletedList(page uint32, pageSize uint32) (fileList []proto.FileInfo, err error) {
	recycledList := db.Model(&fileList).Limit(int(pageSize)).Offset((int(page)-1)*int(pageSize)).Where("trashed IN (?) ", []uint32{proto.TrashStatusPhyDeleted, proto.TrashStatusPhyDelException}).Scan(&fileList)
	err = recycledList.Error
	if err != nil {
		return nil, err
	}
	return
}

func GetSharedCntByBEtag(betag string) (affect int64, err error) {
	var fi []proto.FileInfo
	tx := db.Model(&fi).Where("betag = ?", betag).Scan(&fi)
	return tx.RowsAffected, tx.Error
}

func GetShareBEtagUuids(betag string) (uuids []string, err error) {
	var fi []proto.FileInfo
	err = db.Model(&fi).Select("uuid").Where("betag = ?", betag).Scan(&uuids).Error
	if err != nil {
		return nil, err
	}
	return uuids, nil
}

func GetFolderInfoByUuid(uuid string) (folderInfo proto.FolderInfo, err error) {
	res := db.Model(proto.FileInfo{}).Select("name,path,size,operation_time").Where("uuid = ?", uuid).First(&folderInfo)
	err = res.Error
	if err != nil {
		return folderInfo, err
	} else {
		return folderInfo, nil
	}

}

func GetAllFileInFolder(userId proto.UserIdType, uuid string) (uuids []string, err error) {
	if folderInfo, err := GetFolderInfoByUuid(uuid); err != nil {
		return nil, err
	} else {
		absPath := folderInfo.Path + folderInfo.FolderName + "/"
		db.Model(&proto.FileInfo{}).Where("trashed = 0 AND user_id = ? AND path LIKE ?", userId, absPath+"%").Select("uuid").Scan(&uuids)
		if len(uuids) > 0 {
			return uuids, nil
		}
		return nil, err
	}

}

func PageTotal(userId proto.UserIdType, uuid string, category string, pageSize uint32, trash bool) (uint32, int64, error) {
	var total int64
	var pageNum uint32
	if uuid != "" {
		var subQuery proto.FileInfo
		db.Model(&subQuery).Where("uuid = ?", uuid).First(&subQuery)
		parentPath := subQuery.Path + subQuery.Name + "/"
		db.Model(&proto.FileInfo{}).Where("path = ? AND trashed = ? AND user_id = ?", parentPath, 0, userId).Count(&total)
	} else if category != "" && strings.Contains(category, ",") {
		categories := strings.Split(category, ",")
		err := db.Model(&proto.FileInfo{}).Where("category IN (?,?) AND user_id = ? AND trashed = ?", categories[0], categories[1], userId, 0).Count(&total).Error
		if err != nil {
			return 0, 0, err
		}
	} else if category != "" {
		err := db.Model(&proto.FileInfo{}).Where("category = ? AND user_id = ? AND trashed = ?", category, userId, 0).Count(&total).Error
		if err != nil {
			return 0, 0, err
		}
	} else if trash {
		err := db.Model(&proto.FileInfo{}).Where("trashed = ? AND user_id = ?", 1, userId).Count(&total).Error
		if err != nil {
			return 0, 0, err
		}
	} else {
		err := db.Model(&proto.FileInfo{}).Where("path = ? AND trashed = ? AND user_id = ?", "/", 0, userId).Count(&total).Error
		if err != nil {
			return 0, 0, err
		}
	}
	pageNum = uint32(total) / pageSize
	if uint32(total)%pageSize != 0 {
		pageNum++
	}
	return pageNum, total, nil

}

func SearchPageTotal(userId proto.UserIdType, currentDir string, name string, category string, order string, pageSize uint32) (uint32, int64, error) {
	var total int64
	var pageNum uint32

	if currentDir == "" && category == "" {
		err := db.Model(&proto.FileInfo{}).
			Where("name ILIKE ? AND trashed = ? AND user_id = ?", "%"+"\\"+name+"%", 0, userId).
			Order(order).Count(&total).Error
		if err != nil {
			return 0, 0, err
		}
	} else if category == "" {
		parentPath, folderName := GetPathByUuid(currentDir)
		currentPath := parentPath + folderName + "/"
		err := db.Model(&proto.FileInfo{}).
			Where("path = ? AND name ILIKE ? AND trashed = ? AND user_id = ?", currentPath, "%"+"\\"+name+"%", 0, userId).
			Order(order).Count(&total).Error
		if err != nil {
			return 0, 0, err
		}
	} else if category != "" && strings.Contains(category, ",") {
		categories := strings.Split(category, ",")
		err := db.Model(&proto.FileInfo{}).
			Where("name ILIKE ? AND trashed = ? AND category IN (?,?) AND user_id = ?", "%"+"\\"+name+"%", 0, categories[0], categories[1], userId).
			Order(order).Count(&total).Error
		if err != nil {
			return 0, 0, err
		}
	} else if category != "" {
		err := db.Model(&proto.FileInfo{}).
			Where("name ILIKE ? AND trashed = ? AND category = ? AND user_id = ?", "%"+"\\"+name+"%", 0, category, userId).
			Order(order).Count(&total).Error
		if err != nil {
			return 0, 0, err
		}
	}
	pageNum = uint32(total) / pageSize
	if uint32(total)%pageSize != 0 {
		pageNum++
	}
	return pageNum, total, nil
}

func SyncFolderIsExist(deviceId string, userId proto.UserIdType) bool {
	var deviceInfo proto.SyncInfo
	//var folderInfo proto.FolderInfo
	if deviceId != "" {
		rows := db.Model(&proto.SyncInfo{}).Where("device_id = ? AND user_id = ?", deviceId, userId).First(&deviceInfo).RowsAffected
		if rows != 0 {
			return true
		}
	}
	return false
}

func GetSyncFolderInfo(deviceId string, userId proto.UserIdType) (syncFolderInfo proto.SyncFolderRsp, err error) {
	var info proto.SyncInfo
	var fileInfo proto.FileInfo
	err = db.Model(&proto.SyncInfo{}).Where("device_id = ? AND user_id = ?", deviceId, userId).First(&info).Error
	if err == nil {
		res := db.Model(proto.FileInfo{}).Where("uuid = ?", info.FolderId).First(&fileInfo)
		err = res.Error
		if err == nil && fileInfo.Trashed == 0 {
			db.Model(proto.FileInfo{}).Select("uuid,name,path,user_id").Where("uuid = ?", info.FolderId).First(&syncFolderInfo)
			return syncFolderInfo, nil
		} else if fileInfo.Trashed != 0 {
			return syncFolderInfo, errors.New("sync folder is deleted")
		}
	}
	return
}

func GetSyncedFiles(deviceId string, path string, timestamp uint64, userId proto.UserIdType) (fileInfo proto.FileInfoLst, err error) {
	var folder proto.SyncInfo
	var syncFolder proto.FileInfo
	db.Model(&proto.SyncInfo{}).Where("device_id = ?", deviceId).First(&folder)
	db.Model(&proto.FileInfo{}).Where("uuid = ?", folder.FolderId).First(&syncFolder)
	if path == "" {
		path = "/"
	}

	if path != "" {
		changedList := db.Model(&fileInfo).Where("operation_time > ? AND user_id = ? AND path LIKE ?", timestamp, userId, path+"%").Scan(&fileInfo)
		if changedList.Error != nil {
			return nil, changedList.Error
		} else {
			return fileInfo, nil
		}
	}
	return nil, err
}

func GetUsedSpaceByUser(userId proto.UserIdType) (storage int64, err error) {

	err = db.Raw(fmt.Sprintf("select sum(size) from (select distinct(betag),size from \"aofs_file_infos\" "+
		"where is_dir = false AND user_id = %d AND trashed in (%d,%d,%d)) as subQuery ",
		userId, proto.TrashStatusNormal, proto.TrashStatusLogicDeleted,
		proto.TrashStatusSubFilesLogicDeleted)).Pluck("subQuery", &storage).Error
	if err != nil {
		return 0, err
	} else {
		return storage, nil
	}
}

//func GetBackupStatusByBoxId(boxId string) (status uint8,err error) {
//}

func GetAllFileInfo() (fileInfo []proto.FileInfo, err error) {
	err = db.Model(&proto.FileInfo{}).Where("trashed in (0,1) ").Scan(&fileInfo).Error
	if err != nil {
		return nil, err
	} else {
		return fileInfo, nil
	}

}

func GetAllFileInfoByUserId(userId uint8) (fileInfo []proto.FileInfo, err error) {
	err = db.Model(&proto.FileInfo{}).Where("trashed in (0,1,4) AND user_id = ? ", userId).Scan(&fileInfo).Error
	if err != nil {
		return nil, err
	} else {
		return fileInfo, nil
	}

}

func GetAllFileInfoByUserAndTime(userId proto.UserIdType, timestamp int64) (fileInfo []proto.FileInfo, err error) {
	err = db.Model(&proto.FileInfo{}).Where("trashed in (0,1,4) AND user_id = ? AND operation_time < ?", userId, timestamp).Scan(&fileInfo).Error
	if err != nil {
		return nil, err
	} else {
		return fileInfo, nil
	}

}

func GetTxtFileInfo() (fileInfo []proto.FileInfo, err error) {
	err = db.Model(&proto.FileInfo{}).Where("mime = ? or mime = ?", "text/plain", "text/html").Scan(&fileInfo).Error
	if err != nil {
		return nil, err
	} else {
		return fileInfo, nil
	}
}

