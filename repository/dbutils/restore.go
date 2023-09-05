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

	"gorm.io/gorm"
)

func ProcessSameFile(tx *gorm.DB, userId proto.UserIdType, restoreIds []string) error {

	for _, restoreId := range restoreIds {
		fi, _ := GetFileInfoWithUid(userId, restoreId)
		if fi, _ := GetInfoByPath(userId, fi.Path, fi.Name, proto.TrashStatusNormal); fi != nil {
			newName, err := GenIncNameByPath(userId, fi.Path, fi.Name, proto.TrashStatusNormal)
			if err != nil {
				return err
			}
			err = tx.Debug().Model(&proto.FileInfo{}).Where(ScopeUuid(restoreId), ScopeUser(userId)).
				Update("name", newName).Error

			sql := fmt.Sprintf("UPDATE \"aofs_file_infos\" SET path = replace(path,'%s','%s') where user_id = '%d' AND path like '%s' AND trashed = '%d'",
				fi.Name, newName, userId, fi.Path+fi.Name+"%", proto.TrashStatusSubFilesLogicDeleted)
			err = tx.Debug().Exec(sql).Error
			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return nil
}

func RestoreSubFiles(tx *gorm.DB, userId proto.UserIdType, subRestoreIds []string) error {
	if err := tx.Model(&proto.FileInfo{}).Where(ScopeUuids(subRestoreIds), ScopeUser(userId)).
		Updates(map[string]interface{}{"trashed": proto.TrashStatusNormal,
			"transaction_id": 0,
			"operation_time": time.Now().UnixNano() / 1e6}).Error; err != nil {
		return err
	}

	return nil
}

// RestoreFilesFromTrashV2 重构V2版本的从回收站恢复,支持异步
func RestoreFilesFromTrashV2(userId proto.UserIdType, restoreIds []string, subRestoreIds []string, count int, task *async.AsyncTask) (proto.CodeType, error) {
	tx := db.Begin()

	// 恢复时遇到同名文件的处理
	if err := ProcessSameFile(tx, userId, restoreIds); err != nil {
		logdb.LogE().Err(err).Msg("ProcessSameFile error")
		return 0, err
	}
	// 先把传入的文件uuid恢复
	err := tx.Model(&proto.FileInfo{}).Where(ScopeUser(userId), ScopeUuids(restoreIds)).
		Updates(map[string]interface{}{"trashed": proto.TrashStatusNormal,
			"transaction_id": 0,
			"operation_time": time.Now().UnixNano() / 1e6}).Error


	for _, restoreId := range restoreIds {
		// 恢复时父目录被删除的特殊场景，重建父目录
		fi, _ := GetFileInfoWithUid(userId, restoreId)
		if row := db.Model(&proto.FileInfo{}).Where(ScopeUuid(fi.ParentUuid), ScopeNormalFile()).Find(&proto.FileInfo{}).RowsAffected; row == 0 {
			pathInfo, err := RecursiveCreateFolder(userId, fi.Path)
			if err != nil {
				tx.Rollback()
				logdb.LogE().Err(err).Msg("RecursiveCreateFolder error")
				break
			}
			err = tx.Model(&proto.FileInfo{}).Where(ScopeUser(userId), ScopeUuid(restoreId)).
				Update("parent_uuid", pathInfo.Id).Error
			if err != nil {
				tx.Rollback()
				break
			}
		}
	}

	if err != nil {
		tx.Rollback()
		return 0, err
	}
	if count > env.ASYNC_TASK_THRESHOLD && len(subRestoreIds) > 20 {
		batchSize := len(subRestoreIds) / 20
		batches := (len(subRestoreIds) + batchSize - 1) / batchSize
		// 处理文件夹下的子文件
		go func() {
			task.UpdateStatus(async.AsyncTaskStatusProcessing)
			for i := 1; i <= batches; i++ {
				start := (i - 1) * batchSize
				end := i * batchSize
				//fmt.Println(start,end)
				if end > len(subRestoreIds) {
					end = len(subRestoreIds)
				}
				logdb.LogD().Int("start", start).Int("end", end).Msg("async processing restore")
				logdb.LogD().Msg(fmt.Sprintf("print sub uuids :%v", subRestoreIds[start:end]))
				err = RestoreSubFiles(tx, userId, subRestoreIds[start:end])
				if err != nil {
					task.UpdateStatus(async.AsyncTaskStatusFailed)
					//fmt.Println(err)
					logdb.LogD().Err(err).Msg("RestoreSubFiles error")
					tx.Rollback()
					return
				}

				task.Processed = end
				//fmt.Println(task)
			}

			tx.Commit()
			task.UpdateStatus(async.AsyncTaskStatusSuccess)

		}()
		return proto.CodeCreateAsyncTaskSuccess, nil
	} else {
		err = RestoreSubFiles(tx, userId, subRestoreIds)
		if err != nil {
			logdb.LogD().Err(err).Msg("RestoreSubFiles error")
			tx.Rollback()
			return 0, err
		}

		tx.Commit()
		return proto.CodeOk, nil
	}
}
