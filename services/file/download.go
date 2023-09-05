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

package file

import (
	"aofs/repository/dbutils"
	"aofs/repository/storage"
	"archive/zip"
	"errors"
	"io"
	"os"
)

// ProcessPicZip 批量压缩
func ProcessPicZip(bucket string, uuids []string, w io.Writer, picType int) error {

	if len(uuids) <= 0 {
		return errors.New("zipFile: uuid param error")
	}

	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()
	if picType == 0 {
		for _, fileId := range uuids {
			if err := zipThumb(bucket, fileId, zipWriter); err != nil {
				continue
			}
		}
	} else if picType == 1 {
		for _, fileId := range uuids {
			if err := zipCompressed(bucket, fileId, zipWriter); err != nil {
				continue
			}
		}
	}

	return nil

}

// ZipThumb 压缩缩略图
func zipThumb(bucket string, uuid string, zw *zip.Writer) error {
	fileInfo, err := dbutils.GetInfoByUuid(uuid)
	if err != nil {
		return err
	}

	previewStore := storage.NewPreview()
	path, err := previewStore.GetThumbnailPath(fileInfo.BETag)
	if err != nil {
		return err
	}


	if _, err = os.Stat(path); err == nil {
		fileOnZip, err := zw.Create(fileInfo.Id)
		if err != nil {
			return err
		}
		//读取需要压缩的文件
		subfile, _ := os.Open(path)
		defer subfile.Close()
		_, err = io.Copy(fileOnZip, subfile)
		if err != nil {
			return err
		}
	}

	return nil
}

// ZipThumb 压缩图
func zipCompressed(bucket string, uuid string, zw *zip.Writer) error {
	fileInfo, err := dbutils.GetInfoByUuid(uuid)
	if err != nil {
		return err
	}
	previewStore := storage.NewPreview()

	path, err := previewStore.GetCompressedImgPath(fileInfo.BETag)
	if err != nil {
		return err
	}

	//在压缩包中创建一个文件
	if _, err = os.Stat(path); err == nil {
		fileOnZip, err := zw.Create(fileInfo.Id)
		if err != nil {
			return err
		}
		//读取需要压缩的文件
		subfile, _ := os.Open(path)
		defer subfile.Close()
		_, err = io.Copy(fileOnZip, subfile)
		if err != nil {
			return err
		}
	}

	return nil
}
