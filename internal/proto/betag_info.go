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

type BETagInfo struct {
	BETag      string `gorm:"column:betag;PRIMARY_KEY;index" json:"betag" form:"betag" `
	VolId      uint16 `gorm:"column:vol_id" json:"volId" form:"volId" `
	CreateTime int64  `gorm:"column:created_time" json:"createdAt" form:"createdAt"`
	ModifyTime int64  `gorm:"column:modify_time" json:"modifyAt" form:"modifyAt"`
}

func (BETagInfo) TableName() string {
	return "aofs_betag_infos"
}
