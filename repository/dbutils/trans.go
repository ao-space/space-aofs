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
	"time"

	"gorm.io/gorm"
)

//此文件定义数据事务使用接口
type ITranser interface {
	Finish(err error) error //结束或回滚事务
	TryRollback() error
	Commit() error

	Err() error //获取错误
	MoveFileToTrash(userId proto.UserIdType, deleteId string) (affect int, err error)
	GetInfoByPath(userId proto.UserIdType, path string, name string, trashed uint32) (*proto.FileInfo, error)
	GenIncNameByPath(userId proto.UserIdType, path string, name string, trashed uint32) (string, error)
	AddFileV2(info proto.FileInfo, uuid string) (err error)
	GetSetting(key string) (string, error)
	SetSetting(key string, value string) error
}

type TransProducter interface {
	New() (ITranser, error)
}

type transProduct struct {
}

func (transProduct) New() (ITranser, error) {
	tr := &trans{}
	tr.tx = db.Begin()
	if tr.tx.Error != nil {
		return nil, tr.tx.Error
	}
	return tr, nil
}

func newTrans() (*trans, error) {
	tr := &trans{}
	tr.tx = db.Begin()
	if tr.tx.Error != nil {
		return nil, tr.tx.Error
	}
	return tr, nil
}

func NewTransProducter() TransProducter {
	return transProduct{}
}

type trans struct {
	tx     *gorm.DB
	isOver bool
}

func (tr *trans) Err() error {
	return tr.tx.Error
}

// 收尾
func (tr *trans) Finish(err error) error {
	if tr.isOver {
		return nil
	}

	if tr.Err() != nil || err != nil {
		tr.tx.Rollback()
		return nil
	} else {
		return tr.tx.Commit().Error
	}
}

func (tr *trans) TryRollback() error {
	if tr.isOver {
		return nil
	}
	tr.isOver = true
	return tr.tx.Rollback().Error
}

func (tr *trans) Commit() error {
	if tr.isOver {
		return fmt.Errorf("trans is over")
	}
	tr.isOver = true
	return tr.tx.Commit().Error
}

func (tr *trans) MoveFileToTrash(userId proto.UserIdType, deleteId string) (affect int, err error) {

	//fmt.Println(id)
	// 如果回收站中已存在同名文件则对该文件改名
	var subQuery proto.FileInfo
	var subUuids []string
	var res *gorm.DB
	tr.tx.Model(&proto.FileInfo{}).Where("user_id = ? AND uuid = ?", userId, deleteId).First(&subQuery)
	if _, err := tr.GetInfoByPath(userId, subQuery.Path, subQuery.Name, 1); err == nil {
		// 回收站同名文件改名
		tr.tx.Model(&proto.FileInfo{}).Where("user_id = ? AND path = ? AND name = ? AND trashed = ?", userId, subQuery.Path, subQuery.Name, 1).Update("name", subQuery.Name+time.Now().Format("2006-01-02 15:04:05"))
		tr.tx.Model(&proto.FileInfo{}).Where("user_id = ? AND path LIKE ? AND trashed = ?", userId, subQuery.Path+subQuery.Name+"/%", 4).Update("path", subQuery.Path+subQuery.Name+time.Now().Format("2006-01-02 15:04:05")+"/")
		// 再删除
		if subQuery.IsDir {
			var deleteFolder proto.FileInfo
			tr.tx.Model(&proto.FileInfo{}).Where("user_id = ? AND uuid =?", userId, deleteId).Updates(map[string]interface{}{"trashed": proto.TrashStatusLogicDeleted, "transaction_id": time.Now().Unix(), "operation_time": time.Now().UnixNano() / 1e6})
			tr.tx.Model(&proto.FileInfo{}).Where("user_id = ? AND uuid =?", userId, deleteId).First(&deleteFolder)
			res = tr.tx.Model(&proto.FileInfo{}).Where("user_id = ? AND path LIKE ?", userId, subQuery.Path+subQuery.Name+"/%").Updates(map[string]interface{}{"trashed": proto.TrashStatusSubFilesLogicDeleted, "transaction_id": deleteFolder.TransactionId, "operation_time": time.Now().UnixNano() / 1e6})
			tr.tx.Model(&proto.FileInfo{}).Select("uuid").Where("user_id = ? AND path LIKE ? AND category IN (?,?)", userId, subQuery.Path+subQuery.Name+"/%", "video", "picture").Find(&subUuids)
			affect = int(res.RowsAffected + 1)
		} else {

			res = tr.tx.Model(&proto.FileInfo{}).Where("user_id = ? AND uuid = ?", userId, deleteId).Updates(map[string]interface{}{"trashed": 1, "transaction_id": time.Now().Unix(), "operation_time": time.Now().UnixNano() / 1e6})
			affect = 1
		}

		err = res.Error

	} else {
		if subQuery.IsDir {
			var deleteFolder proto.FileInfo
			tr.tx.Model(&proto.FileInfo{}).Where("user_id = ? AND uuid =?", userId, deleteId).Updates(map[string]interface{}{"trashed": proto.TrashStatusLogicDeleted, "transaction_id": time.Now().Unix(), "operation_time": time.Now().UnixNano() / 1e6})
			tr.tx.Model(&proto.FileInfo{}).Where("user_id = ? AND uuid =?", userId, deleteId).First(&deleteFolder)
			res = tr.tx.Model(&proto.FileInfo{}).Where("path LIKE ? AND user_id = ?", subQuery.Path+subQuery.Name+"/%", userId).Updates(map[string]interface{}{"trashed": proto.TrashStatusSubFilesLogicDeleted, "transaction_id": deleteFolder.TransactionId, "operation_time": time.Now().UnixNano() / 1e6})
			tr.tx.Model(&proto.FileInfo{}).Select("uuid").Where("user_id = ? AND path LIKE ?", userId, subQuery.Path+subQuery.Name+"/%").Find(&subUuids)

			affect = int(res.RowsAffected + 1)
		} else {

			res = tr.tx.Model(&proto.FileInfo{}).Where("user_id = ? AND uuid =?", userId, deleteId).Updates(map[string]interface{}{"trashed": 1, "transaction_id": time.Now().Unix(), "operation_time": time.Now().UnixNano() / 1e6})
			affect = 1
		}
	}

	if err != nil {
		return 0, err
	} else {
		return affect, nil
	}

}
func (tr *trans) GetInfoByUuid(uuid string) (*proto.FileInfo, error) {
	var fileInfo proto.FileInfo
	tx := tr.tx.Model(fileInfo).Where("uuid = ?", uuid).First(&fileInfo)
	if tx.Error != nil {
		return nil, tr.tx.Error
	} else if tx.RowsAffected == 0 {
		return nil, fmt.Errorf("record not exist")
	} else {
		return &fileInfo, nil
	}
}

func (tr *trans) AddFile(info proto.FileInfo) (err error) {
	err = tr.tx.Model(&proto.FileInfo{}).Create(&info).Error
	return
}

// AddFileV2 上传文件,path 传入参数为当前目录的uuid
func (tr *trans) AddFileV2(info proto.FileInfo, uuid string) (err error) {
	fi, err := tr.GetInfoByUuid(uuid)
	if err != nil {
		return err
	}

	info.Path = fi.AbsPath()
	return tr.AddFile(info)
}

func (tr *trans) GetInfoByPath(userId proto.UserIdType, path string, name string, trashed uint32) (*proto.FileInfo, error) {
	var fileInfo proto.FileInfo
	tx := tr.tx.Model(fileInfo).Where("user_id = ? AND name = ? AND path = ? AND trashed = ?", userId, name, path, trashed).First(&fileInfo)
	if tx.Error == nil {
		return &fileInfo, nil
	} else { //if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return nil, tx.Error
	}

}

func (tr *trans) GenIncNameByPath(userId proto.UserIdType, path string, name string, trashed uint32) (string, error) {

	_, err := tr.GetInfoByPath(userId, path, name, trashed)
	if err == nil {
		//按规则处理
		ext := filepath.Ext(name)
		preName := name[:len(name)-len(ext)]
		for i := 1; ; i++ {
			newName := preName + fmt.Sprintf("(%d)", i) + ext
			_, err = tr.GetInfoByPath(userId, path, newName, trashed)
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

func (tr *trans) hisTaskOk(taskName string) error {
	var htf proto.Setting
	result := tr.tx.Model(&proto.Setting{}).Where("setting_name = ?", proto.HIS_TASK_BETAG).First(&htf)
	if result.Error != nil && result.Error == gorm.ErrRecordNotFound {
		return result.Error
	}
	if htf.Value == proto.HIS_TASK_STATUS_OK {
		return nil
	} else {
		return fmt.Errorf(htf.Value) //这个流程应该走不了，因为成功了才会被记录
	}
}

func (tr *trans) GetSetting(key string) (string, error) {
	var htf proto.Setting
	result := tr.tx.Model(&proto.Setting{}).Where("setting_name = ?", key).First(&htf)
	if result.Error == nil {
		return htf.Value, nil
	} else {
		return "", result.Error
	}
}

func (tr *trans) SetSetting(key string, value string) error {
	var htf proto.Setting
	htf.Name = key
	htf.Value = value
	htf.CreateTime = time.Now().Unix()
	result := tr.tx.Model(proto.Setting{}).Create(&htf)
	return result.Error
}

func (tr *trans) getBETagInfo(betag string) (*proto.BETagInfo, error) {
	var bi proto.BETagInfo
	result := tr.tx.Model(proto.BETagInfo{}).Where("betag=?", betag).First(&bi)
	if result.Error != nil {
		return nil, result.Error
	} else {
		return &bi, nil
	}
}

func (tr *trans) deleteBETagInfo(betag string) error {
	result := tr.tx.Delete(proto.BETagInfo{}, "betag=?", betag)
	return result.Error
}
func (tr *trans) addBETagInfo(betag string, volId int) error {
	var bi proto.BETagInfo
	bi.BETag = betag
	bi.VolId = uint16(volId)
	bi.CreateTime = time.Now().Unix()
	bi.ModifyTime = time.Now().Unix()
	result := tr.tx.Model(proto.BETagInfo{}).Create(&bi)
	return result.Error

}
