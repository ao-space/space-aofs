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

//--- 分页消息结构定义
type PageInfo struct {
	Page     uint32 `json:"page" form:"page"`
	PageSize uint32 `json:"pageSize" form:"pageSize"`
}

type PageInfoExt struct {
	PageInfo
	FileCount int64  `json:"count" form:"count"`
	TotalPage uint32 `json:"total" form:"total"`
}
