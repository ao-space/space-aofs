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

/*
 * @Author: zhongguang
 * @Date: 2022-12-09 16:12:37
 * @Last Modified by: zhongguang
 * @Last Modified time: 2022-12-09 16:23:20
 */

package storage

import (
	"aofs/internal/proto"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type MultiDisker interface {
	GetRootPath(diskId int) (string, error)
}

type Indexer interface {
	Get(key string) (int, error)      //获取索引信息
	Delete(key string) (int, error)   //删除索引
	Add(key string, diskId int) error //增加 1 条索引
}

type Bucketer interface {
	GetDiskid(key string) (bool, error)
	//Put(key string, r io.Reader, attrs map[string]interface{}) error
	Get(key string, part *proto.Part) (io.ReadCloser, error)
	Del(key string) error
	GetPath(key string) (string, error)
	GetRelativePath(key string) string
	GetMultipartPath() string //获取分片上传集中存储路径
	GetFileAbsPath(key string) string
}

type bucket struct {
	disker  MultiDisker
	indexer Indexer
	name    string //桶的名称，即子路径的名称
}

func (dp *bucket) IsExist(key string) bool {
	fpath := dp.getFilePath(dp.name, key)
	_, err := os.Stat(fpath)
	return err == nil
}

func (dp *bucket) getFilePath(rootPath string, key string) string {
	if len(key) > 4 {
		h1 := key[:2]
		h2 := key[2:4]
		return filepath.Join(rootPath, dp.name, h1, h2, key)
	}
	return fmt.Sprintf("file's betag error :%v", io.EOF)

}

func (dp *bucket) getDir(rootPath string, key string) (string, error) {
	if len(key) > 4 {
		h1 := key[:2]
		h2 := key[2:4]

		dir := filepath.Join(rootPath, dp.name, h1, h2)
		err := os.MkdirAll(dir, os.ModePerm)
		return dir, err
	}
	return "", io.EOF

}

func (bkt *bucket) GetDiskid(key string) (int, error) {
	return bkt.indexer.Get(key)
}

func (bkt *bucket) Get(key string, part *proto.Part) (io.ReadCloser, error) {
	diskId, err := bkt.indexer.Get(key)
	if err != nil {
		return nil, err
	}

	rootPath, err := bkt.disker.GetRootPath(diskId)
	if err != nil {
		return nil, err
	}

	filepath := bkt.getFilePath(rootPath, key)

	if file, err := os.Open(filepath); err != nil {
		return file, err
	} else if part == nil {
		return file, err
	} else {
		if _, err := file.Seek(part.Start, os.SEEK_SET); err != nil {
			file.Close()
			return nil, err
		}

		if part.End == -1 {
			return file, nil
		}

		if stat, err := file.Stat(); err != nil {
			file.Close()
			return nil, err
		} else if part.End >= stat.Size() {
			file.Close()
			return nil, fmt.Errorf("Range Not Satisfiable")
		} else {
			r := io.LimitReader(file, part.Len())
			type RC struct {
				io.Reader
				io.Closer
			}
			var rc RC
			rc.Reader = r
			rc.Closer = file
			return rc, nil
		}
	}

}

func (bkt *bucket) Del(key string) error {

	diskId, err := bkt.indexer.Get(key)
	if err != nil {
		return err
	}

	rootPath, err := bkt.disker.GetRootPath(diskId)
	if err != nil {
		return err
	}

	fpath := bkt.getFilePath(rootPath, key)

	err = os.Remove(fpath)

	logger.LogI().Msg(fmt.Sprintf("push deleteMsg to redis,key:%v", key))
	PushMsg(map[string]interface{}{"key": key, "bucket": bkt.name}, "delete")
	if err != nil && os.IsNotExist(err) {
		logger.LogE().Msg(fmt.Sprintf("Remove files err:%v", err))
		return nil
	}
	return err
}
