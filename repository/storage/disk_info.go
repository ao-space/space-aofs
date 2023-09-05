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

package storage

import (
	"aofs/internal/env"
	"aofs/internal/utils"
	"fmt"
	"path/filepath"
)

type SharedDiskInfo struct {
	DiskInitialCode     int    `json:"diskInitialCode"`     // 1: 磁盘正常; 2: 未初始化;3: 正在格式化; 4: 正在数据同步; 100: 未知错误; 101:  磁盘格式化错误; >101: 其他初始化错误;
	DiskInitialMessage  string `json:"diskInitialMessage"`  // 磁盘初始化结果/异常信息。
	DiskInitialProgress uint   `json:"diskInitialProgress"` // 磁盘初始化进度。

	DiskExpandCode     int    `json:"diskExpandCode"`     // 1: 扩容完成; 2: 未扩容状态; 3:正在扩容; 100: 扩容未知错误; 101:  扩容磁盘格式化错误; >101: 扩容其他错误;
	DiskExpandMessage  string `json:"diskExpandMessage"`  // 磁盘扩容结果/异常信息。
	DiskExpandProgress uint   `json:"diskExpandProgress"` // 磁盘扩容进度。

	CreatedTime string `json:"createdTime"` // 创建时间
	UpdatedTime string `json:"updatedTime"` // 更新时间

	RaidType              int      `json:"raidType"`              // 1: normal; 2: raid1。
	RaidDiskHwIds         []string `json:"raidDiskHwIds"`         // 参与 raid 的磁盘 id 列表.
	PrimaryStorageHwIds   []string `json:"PrimaryStorageHwIds"`   // 主存储磁盘硬件 id 列表
	SecondaryStorageHwIds []string `json:"secondaryStorageHwIds"` // 次存储磁盘硬件 id 列表

	DiskMountInfos []*DiskMountInfo `json:"diskMountInfos"`

	FileStorageVolumePathPrefix string `json:"fileStorageVolumePathPrefix"` // 挂载给 aofs 容器的目录 /home/eulixspace_file_storage/parts/bp_part_nvme_3841a657 中 "bp_part_" 这样的前缀.
}

type DiskMountInfo struct {
	HwIds                []string `json:"hwIds"`               // e.g. "hwIds": ["52fa9e6c", "bd2c7570" ], 格式化之后不会变化
	MountDevice          string   `json:"mountDevice"`         // e.g. /dev/sda1, /dev/sdb1, /dev/vg0/lv0
	DeviceUuid           string   `json:"deviceUuid"`          // 格式化之后会变化
	DeviceSequenceNumber int64    `json:"dviceSequenceNumber"` // 设备自增序列号, 根据 DeviceUuid 判断，不重复时算新磁盘。从1自增。
	MountPath            string   `json:"mountPath"`           // e.g. /mnt/bp/data/raid1_52fa9e6c-bd2c7570/mountpoint
	DataFolderRoot       string   `json:"dataFolderRoot"`      // raid1_138cf704-52fa9e6c , nvme_3841a657 in /mnt/bp/data/raid1_138cf704-52fa9e6c/mountpoint, /mnt/bp/data/nvme_3841a657/mountpoint"
	MapperName           string   `json:"mapperName"`          // e.g. luks_raid1_52fa9e6c-bd2c7570
	FSType               string   `json:"fSType"`              // ext4
	IsPrimaryStorage     bool     `json:"isPrimaryStorage"`    // 是否是主存储
}

func (di *DiskMountInfo) getPath(prefix string) string {
	return filepath.Join(env.DATA_PATH, fmt.Sprintf("%v%v", prefix, di.DataFolderRoot))
}

func getDiskInfo() (*SharedDiskInfo, error) {
	var sdi SharedDiskInfo
	if err := utils.ReadJsonFromFile(filepath.Join(env.SHARED_PATH, "disk_info.json"), &sdi); err != nil {
		return nil, err
	}
	if len(sdi.DiskMountInfos) == 0 {
		return nil, fmt.Errorf("not found disk mount info")
	}

	return &sdi, nil

}
