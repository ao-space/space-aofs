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
	"aofs/internal/proto"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"time"
)

// type Indexer interface {
// 	Get(key string) (int, error)      //获取索引信息
// 	Delete(key string) (int, error)   //删除索引
// 	Add(key string, diskId int) error // 增加1条索引
// }

type MockIndexer struct {
	mapDiskFile map[string]int
}

func (mi *MockIndexer) Get(key string) (int, error) {
	if diskId, ok := mi.mapDiskFile[key]; ok {
		return diskId, nil
	}
	return 0, fmt.Errorf("not found")
}
func (mi *MockIndexer) Delete(key string) (int, error) {
	delete(mi.mapDiskFile, key)
	return 1, nil
}

func (mi *MockIndexer) Add(key string, diskId int) error {
	mi.mapDiskFile[key] = diskId
	return nil
}

func TestPutAndGet(t *testing.T) {
	var mi MockIndexer
	mi.mapDiskFile = map[string]int{}

	var md multiDisk
	if err := md.Init(&mi); err != nil {
		t.Error(err)
	}

	text := fmt.Sprintf("%v", time.Now().Unix())
	bucket := "bucketa"
	key := text

	if err := md.Put(bucket, key, strings.NewReader(text), int64(len(text))); err != nil {
		t.Error(err)
	} else {
		if rc, err := md.Get(bucket, key, &proto.Part{Start: 0, End: 2}); err != nil {
			t.Error(err)
		} else {
			if data, err := ioutil.ReadAll(rc); err != nil {
				t.Error(err)
			} else if string(data) != text[0:3] {
				t.Error(fmt.Errorf("data err:[%v]", string(data)))
			}
		}
	}
}
