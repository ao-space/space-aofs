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

type SyncDeviceReq struct {
	DeviceId   string `json:"deviceId" form:"deviceId" validate:"required"`
	DeviceName string `json:"deviceName" form:"deviceName" validate:"required"`
}

type SyncFolderRsp struct {
	FolderId   string `gorm:"column:uuid" json:"uuid" form:"uuid"`
	FolderName string `gorm:"column:name" json:"name" form:"name"`
	FolderPath string `gorm:"column:path" json:"path" form:"path"`
	UserId     uint8  `gorm:"column:user_id" json:"userId" form:"userId"`
}

type SyncedFilesReq struct {
	Timestamp uint64 `json:"timestamp" form:"timestamp" validate:"required,gte=0"`
	DeviceId  string `json:"deviceId" form:"deviceId"`
	Path      string `json:"path" form:"path"`
}
