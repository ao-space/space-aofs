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

package main

type Monitor struct {
	Alloc        uint64 `json:"堆内存"`   // 堆内存字节分配数
	TotalAlloc   uint64 `json:"最大堆内存"` // 堆中分配最大字节数
	Sys          uint64 `json:"系统内存"`  // 从系统中获得的总内存
	Mallocs      uint64 `json:"分配对象"`  // 分配对象数
	Frees        uint64 `json:"释放对象"`  // 释放对象数
	LiveObjects  uint64 `json:"存活对象"`  // 存活对象数
	PauseTotalNs uint64 `json:"GC时间"`  // 总GC暂停时间
	LastGC       uint64 `json:"最后一次GC时间戳"`

	NumGC        uint32 `json:"GC次数"`
	NumGoroutine int    `json:"协程数"` // goroutine数量
}