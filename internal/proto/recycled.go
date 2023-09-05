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

type RestoreRecycledReq struct {
	RecycledUuids []string `json:"uuids" form:"uuids" validate:"required,dive,uuid"`
}

type UuidLst struct {
	Uuids []string `json:"uuids" form:"uuids"`
}

//-------- 清理回收站的文件或文件夹

//请求消息
type RecycledPhyDeleteReq = UuidLst //移除文件或文件夹的请求

//回应消息（无包体，只有包头）
