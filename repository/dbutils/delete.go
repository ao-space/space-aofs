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
	"aofs/internal/env"
	"aofs/internal/proto"
	"aofs/services/async"
	"fmt"
	"time"

	"path"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//const AsyncTaskThreshold = 1000

func ProcessSameFileInTrash(tx *gorm.DB, userId proto.UserIdType, deleteIds []string) error {
	for _, deleteId := range deleteIds {
		subQuery, err := GetInfoByUuid(deleteId)
		if err != nil {
			continue
		}
		if _, err = GetInfoByPath(userId, subQuery.Path, subQuery.Name, proto.TrashStatusLogicDeleted); err == nil {
			newFileName := subQuery.Name[:len(subQuery.Name)-len(path.Ext(subQuery.Name))] + time.Now().Format("2006-01-02 15:04:05") + path.Ext(subQuery.Name)
			if err := tx.Model(&proto.FileInfo{}).Where("user_id = ? AND path = ? AND name = ? AND trashed = ?", userId, subQuery.Path, subQuery.Name, 1).Update("name", newFileName).Error; err != nil {
				return err
			}
			newPath := subQuery.Path + subQuery.Name + time.Now().Format("2006-01-02 15:04:05") + "/"
			if err := tx.Model(&proto.FileInfo{}).Where("user_id = ? AND path LIKE ? AND trashed = ?", userId, subQuery.Path+subQuery.Name+"/%", 4).Update("path", newPath).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func MoveFileToTrashV2(userId proto.UserIdType, deleteIds []string, subDeleteIds []string, count int, task *async.AsyncTask) (proto.CodeType, error) {
	tx := db.Begin()

	transactionId := time.Now().Unix()

	// 回收站有同名文件处理
	if err := ProcessSameFileInTrash(tx, userId, deleteIds); err != nil {
		return 0, err
	}
	// 先把传入的文件uuid 删除
	err := tx.Model(&proto.FileInfo{}).
		Where(ScopeUser(userId), ScopeUuids(deleteIds)).
		Updates(map[string]interface{}{"trashed": proto.TrashStatusLogicDeleted, "transaction_id": transactionId, "operation_time": time.Now().UnixNano() / 1e6}).Error

	logdb.LogD().Msg(fmt.Sprintf("process delete %d files", count))
	if count > env.ASYNC_TASK_THRESHOLD && len(subDeleteIds) > 20 {

		batchSize := len(subDeleteIds) / 20
		batches := (len(subDeleteIds) + batchSize - 1) / batchSize
		// 处理文件夹下的子文件
		go func() {
			
			task.UpdateStatus(async.AsyncTaskStatusProcessing)
			for i := 1; i <= batches; i++ {
				start := (i - 1) * batchSize
				end := i * batchSize
				
				if end > len(subDeleteIds) {
					end = len(subDeleteIds)
				}

				logdb.LogD().Int("start", start).Int("end", end).Msg("async processing delete")
				
				err = DeleteSubFilesBackend(tx, userId, subDeleteIds[start:end], transactionId)
				if err != nil {
					task.UpdateStatus(async.AsyncTaskStatusFailed)
					//fmt.Println(err)
					logdb.LogD().Err(err).Msg("DeleteSubFilesBackend error")
					tx.Rollback()
					return
				}

				task.Processed = end
				
			}

			tx.Commit()
			task.UpdateStatus(async.AsyncTaskStatusSuccess)

		}()

		return proto.CodeCreateAsyncTaskSuccess, nil
	} else {
		err = DeleteSubFilesBackend(tx, userId, subDeleteIds, transactionId)
		if err != nil {
			//fmt.Println(err)
			logdb.LogD().Err(err).Msg("DeleteSubFilesBackend error")
			tx.Rollback()
			return 0, err
		}

		return proto.CodeOk, tx.Commit().Error
	}
}

func DeleteSubFilesBackend(tx *gorm.DB, userId proto.UserIdType, deleteIds []string, transactionId int64) error {
	//defer tx.Rollback()
	//var subQuery proto.FileInfo
	if len(deleteIds) == 0 {
		return nil
	}

	if err := tx.Model(&proto.FileInfo{}).Where(ScopeUuids(deleteIds), ScopeUser(userId)).
		Updates(map[string]interface{}{"trashed": proto.TrashStatusSubFilesLogicDeleted, "transaction_id": transactionId, "operation_time": time.Now().UnixNano() / 1e6}).Error; err != nil {
		return err
	}

	return nil
}

// MoveFileToTrash  移动文件到回收站
func MoveFileToTrash(userId proto.UserIdType, deleteId string) (affect int, err error) {
	tx := db.Begin()
	//fmt.Println(id)
	// 如果回收站中已存在同名文件则对该文件改名
	var subQuery proto.FileInfo
	var subUuids []string
	var res *gorm.DB
	tx.Model(&proto.FileInfo{}).Where("user_id = ? AND uuid = ?", userId, deleteId).First(&subQuery)

	if _, err := GetInfoByPath(userId, subQuery.Path, subQuery.Name, 1); err == nil {
		// 回收站同名文件改名
		tx.Model(&proto.FileInfo{}).Where("user_id = ? AND path = ? AND name = ? AND trashed = ?", userId, subQuery.Path, subQuery.Name, 1).Update("name", subQuery.Name+time.Now().Format("2006-01-02 15:04:05"))
		tx.Model(&proto.FileInfo{}).Where("user_id = ? AND path LIKE ? AND trashed = ?", userId, subQuery.Path+subQuery.Name+"/%", 4).Update("path", subQuery.Path+subQuery.Name+time.Now().Format("2006-01-02 15:04:05")+"/")
		// 再删除
		if subQuery.IsDir {
			var deleteFolder proto.FileInfo
			err = tx.Debug().Model(&proto.FileInfo{}).Where("user_id = ? AND uuid =?", userId, deleteId).Updates(map[string]interface{}{"trashed": proto.TrashStatusLogicDeleted, "transaction_id": time.Now().Unix(), "operation_time": time.Now().UnixNano() / 1e6}).Error
			if err != nil {
				return 0, err
			}
			tx.Debug().Model(&proto.FileInfo{}).Where("user_id = ? AND uuid =?", userId, deleteId).First(&deleteFolder)
			res = tx.Debug().Model(&proto.FileInfo{}).Where("user_id = ? AND path LIKE ? AND trashed = 0", userId, subQuery.Path+subQuery.Name+"/%").Updates(map[string]interface{}{"trashed": proto.TrashStatusSubFilesLogicDeleted, "transaction_id": deleteFolder.TransactionId, "operation_time": time.Now().UnixNano() / 1e6})
			tx.Model(&proto.FileInfo{}).Select("uuid").Where("user_id = ? AND path LIKE ? AND category IN (?,?) AND trashed = 4", userId, subQuery.Path+subQuery.Name+"/%", "video", "picture").Find(&subUuids)

			affect = int(res.RowsAffected + 1)
		} else {

			res = tx.Model(&proto.FileInfo{}).Where("user_id = ? AND uuid = ?", userId, deleteId).Updates(map[string]interface{}{"trashed": 1, "transaction_id": time.Now().Unix(), "operation_time": time.Now().UnixNano() / 1e6})
			affect = 1
		}

		err = res.Error

	} else {
		if subQuery.IsDir {
			var deleteFolder proto.FileInfo
			err := tx.Debug().Model(&proto.FileInfo{}).Where("user_id = ? AND uuid =?", userId, deleteId).Updates(map[string]interface{}{"trashed": proto.TrashStatusLogicDeleted, "transaction_id": time.Now().Unix(), "operation_time": time.Now().UnixNano() / 1e6}).Error
			if err != nil {
				return 0, err
			}
			tx.Debug().Model(&proto.FileInfo{}).Where("user_id = ? AND uuid =?", userId, deleteId).First(&deleteFolder)
			res = tx.Debug().Model(&proto.FileInfo{}).Where("path LIKE ? AND user_id = ? AND trashed = 0", subQuery.Path+subQuery.Name+"/%", userId).Updates(map[string]interface{}{"trashed": proto.TrashStatusSubFilesLogicDeleted, "transaction_id": deleteFolder.TransactionId, "operation_time": time.Now().UnixNano() / 1e6})
			tx.Model(&proto.FileInfo{}).Select("uuid").Where("user_id = ? AND path LIKE ? AND category IN (?,?) AND trashed = 4", userId, subQuery.Path+subQuery.Name+"/%", "video", "picture").Find(&subUuids)

			affect = int(res.RowsAffected + 1)
		} else {

			res = tx.Model(&proto.FileInfo{}).Where("user_id = ? AND uuid =?", userId, deleteId).Updates(map[string]interface{}{"trashed": 1, "transaction_id": time.Now().Unix(), "operation_time": time.Now().UnixNano() / 1e6})
			affect = 1
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

// 删除回收站文件或目录， 1-》2 & 4->2
func RecycledFromLogicToPhy(userId proto.UserIdType, uuids []string) (affect int, err error) {

	affect = 0
	err = nil
	if len(uuids) == 0 {
		//清空回收站
		var files []string
		//var duplicates []string
		// totalPath 唯一性约束冲突，更新trashed字段为3
		db.Debug().Model(&proto.FileInfo{}).Where("trashed = ? AND is_dir = ?", proto.TrashStatusLogicDeleted, false).
			Select("betag").Scan(&files)

		for _, file := range files {
			db.Debug().Model(&proto.FileInfo{}).Where("trashed = ? AND is_dir = ? AND betag = ?", proto.TrashStatusSubFilesLogicDeleted, false, file).
				Update("trashed", proto.TrashStatusPhyDelException)
		}
		res := db.Debug().Model(&proto.FileInfo{}).
			Where("trashed in (?) ", []uint32{proto.TrashStatusLogicDeleted, proto.TrashStatusSubFilesLogicDeleted}).
			Clauses(clause.OnConflict{OnConstraint: "totalPath",
				DoUpdates: clause.Assignments(map[string]interface{}{"trashed": 3})}).
			Update("trashed", proto.TrashStatusPhyDeleted)

		affect += int(res.RowsAffected)
		err = res.Error
	} else {
		var fi *proto.FileInfo
		for _, uuid := range uuids {
			if fi, err = GetFileInfoWithUid(userId, uuid); err != nil {
				break
			} else {
				logdb.LogI().Msg(fmt.Sprintf("recycled fileinfo: %v", fi))
				//log.Println("recycled fi:", fi)
				//状态不对，或没有 transid
				if (fi.Trashed != proto.TrashStatusLogicDeleted && fi.Trashed != proto.TrashStatusSubFilesLogicDeleted) || fi.TransactionId <= 0 {
					logdb.LogE().Msg("param invalid")
					err = fmt.Errorf("param invalid")
					break
				}
				var res *gorm.DB
				//根据操作事务id清理之
				if fi.IsDir == false {
					res = db.Debug().Model(&proto.FileInfo{}).Where("transaction_id = ? AND uuid = ?", fi.TransactionId, fi.Id).Update("trashed", proto.TrashStatusPhyDeleted)

				} else {
					path := fi.AbsPath()
					db.Model(&proto.FileInfo{}).Where("transaction_id = ? AND uuid = ?", fi.TransactionId, fi.Id).Update("trashed", proto.TrashStatusPhyDeleted)
					res = db.Model(&proto.FileInfo{}).Where("transaction_id = ? AND path LIKE ?", fi.TransactionId, path+"%").Update("trashed", proto.TrashStatusPhyDeleted)
				}

				affect += int(res.RowsAffected)
				err = res.Error
				if err != nil {
					break
				}

			}
		}
	}

	return affect, nil
}

func RecycledFromPhyToException(uuid string) (affect int, err error) {

	//将记录标记为清理异常
	res := db.Model(&proto.FileInfo{}).Where("uuid = ?", uuid).Update("trashed", proto.TrashStatusPhyDelException)
	affect = int(res.RowsAffected)
	err = res.Error

	return
}

func DeleteByUuid(uuid string) (affect int64, err error) {

	tx := db.Delete(&proto.FileInfo{}, "uuid = ?", uuid)
	affect = tx.RowsAffected
	err = tx.Error
	return
}

func DeleteUser(userId proto.UserIdType) error {
	//var userFiles []proto.FileInfo

	tx := db.Begin()

	err := tx.Model(&proto.FileInfo{}).Where("user_id = ? AND trashed IN (?,?,?) ", userId,
		proto.TrashStatusNormal, proto.TrashStatusLogicDeleted, proto.TrashStatusSubFilesLogicDeleted).
		Updates(map[string]interface{}{"trashed": proto.TrashStatusPhyDeleted, "transaction_id": time.Now().Unix()}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func PhyDeleteByUuid(userId uint8, uuid string) error {
	err := db.Model(&proto.FileInfo{}).Where("user_id = ? AND uuid = ?  ", userId, uuid).
		Updates(map[string]interface{}{"trashed": proto.TrashStatusPhyDeleted, "transaction_id": time.Now().Unix()}).Error
	if err != nil {
		return err
	}
	return nil
}
