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
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type dataPath struct {
	path string
}

func (dp *dataPath) IsExist(bucket string, key string) bool {
	fpath := dp.getFilePath(bucket, key)
	_, err := os.Stat(fpath)
	return err == nil
}

func (dp *dataPath) getFilePath(bucket string, key string) string {
	if len(key) > 4 {
		h1 := key[:2]
		h2 := key[2:4]
		return filepath.Join(dp.path, bucket, h1, h2, key)
	}
	return fmt.Sprintf("file's betag error :%v", io.EOF)

}

func (dp *dataPath) getDir(bucket string, key string) (string, error) {
	if len(key) > 4 {
		h1 := key[:2]
		h2 := key[2:4]

		dir := filepath.Join(dp.path, bucket, h1, h2)
		err := os.MkdirAll(dir, os.ModePerm)
		return dir, err
	}
	return "", io.EOF

}

func GetPathSize(path string) int64 {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return fileInfo.Size()
}
