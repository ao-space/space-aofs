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
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"aofs/internal/env"
	"aofs/internal/proto"
	"aofs/internal/utils"
	"aofs/repository/dbutils"
)

/*
type MultiDisker interface {
	Alloc(size int64) (string, error)
	GetRootPath(diskId int) (string,  error)
}
*/

var ErrEnoughSpace error = errors.New("not Enough Space")

type MultiDiskStorager interface {
	IsExist(bucket string, key string) (bool, error)
	Put(bucket string, key string, r io.Reader, size int64) error
	Get(bucket string, key string, part *proto.Part) (io.ReadCloser, error)
	Del(bucket string, key string) error
	GenPath(bucket string, key string, size int64) (int, string, error) //分配文件实际存储盘
	GetDiskPath(diskId int) (string, error)
	GetDiskPathByBEtag(key string) (string, error)
	GetDiskMPPath(diskId int) (string, error)
	MoveFile(path string, diskId int, bucket, key string) (string, error) //将文件移动到内部
	GetPath(bucket string, key string) (string, error)
	GetRelativePath(bucket string, key string) (string, error)
	GetMultipartPath() (int, string) //获取分片上传集中存储路径
	GetFileAbsPath(bucket string, key string) (string, error)
}

type multiDisk struct {
	mapDisk map[int]string
	indexer Indexer
}

var md multiDisk

//初始化目录
func (m *multiDisk) Init(idx Indexer) error {
	m.indexer = idx
	m.mapDisk = make(map[int]string, 5)

	if sdi, err := getDiskInfo(); err != nil {
		return err
	} else {
		for _, di := range sdi.DiskMountInfos {

			m.mapDisk[int(di.DeviceSequenceNumber)] = di.getPath(sdi.FileStorageVolumePathPrefix)
			m.initDiskDir(m.mapDisk[int(di.DeviceSequenceNumber)], *di)
		}

		dir := filepath.Join(env.DATA_PATH, "multipart-meta")
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			logger.LogF().Err(err).Msg("Failed to init multipart meta dir.")
			return err
		}
	}

	fmt.Println("dirs:", m.mapDisk)
	return nil
}

func (m *multiDisk) GetDiskPath(diskId int) (string, error) {
	if path, ok := m.mapDisk[diskId]; ok {
		return path, nil
	} else {
		return "", fmt.Errorf("not found diskId:%v", diskId)
	}
}

func (m *multiDisk) GetDiskMPPath(diskId int) (string, error) {
	if path, err := m.GetDiskPath(diskId); err != nil {
		return "", err
	} else {
		return filepath.Join(path, "multipart"), nil
	}

}
func (m *multiDisk) MoveFile(path string, diskId int, bucket, key string) (string, error) {
	if dstDir, err := m.PreDir(diskId, bucket, key, true); err != nil {
		return "", err
	} else {
		fpath := filepath.Join(dstDir, key)
		if err := os.Rename(path, fpath); err != nil {
			return "", err
		} else {
			return fpath, m.indexer.Add(key, diskId)
		}
	}
}

func (m *multiDisk) initDiskDir(path string, di DiskMountInfo) error {

	if err := utils.WriteJsonToFile(filepath.Join(path, ".disk.info"), di); err != nil {
		return err
	}

	dir := filepath.Join(path, "multipart")
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		logger.LogF().Err(err).Msg("Failed to init multipart dir.")
		return err
	}

	return nil
}

func (m *multiDisk) getFilePath(bucket string, key string) (string, error) {
	if dir, err := m.getFileDir(bucket, key); err != nil {
		return "", err
	} else {
		return filepath.Join(dir, key), nil
	}
}

func (m *multiDisk) getFileDir(bucket string, key string) (string, error) {
	if len(key) < 4 {
		return "", fmt.Errorf("key(%v) is invalid", key)
	}

	diskId, err := m.indexer.Get(key)
	if err != nil {
		return "", err
	}

	diskPath, err := m.GetDiskPath(diskId)
	if err != nil {
		return "nil", err
	}

	h1 := key[:2]
	h2 := key[2:4]
	return filepath.Join(diskPath, bucket, h1, h2), nil
}

func (m *multiDisk) IsExist(bucket string, key string) (bool, error) {
	if fpath, err := m.getFilePath(bucket, key); err != nil {
		return false, err
	} else {
		_, err := os.Stat(fpath)
		return err == nil, nil
	}

}

func (m *multiDisk) Get(bucket string, key string, part *proto.Part) (io.ReadCloser, error) {

	filepath, err := m.getFilePath(bucket, key)
	if err != nil {
		return nil, err
	}

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
			return nil, fmt.Errorf("range not satisfiable")
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

func (m *multiDisk) PreDir(diskId int, bucket string, key string, mkdir bool) (string, error) {
	if path, err := m.GetDiskPath(diskId); err != nil {
		return "", err
	} else {
		h1 := key[:2]
		h2 := key[2:4]

		dir := filepath.Join(path, bucket, h1, h2)
		if mkdir {
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				return "", err
			}
		}
		return dir, nil
	}
}

func (m *multiDisk) GenPath(bucket string, key string, size int64) (int, string, error) {
	err := fmt.Errorf("failed to gen path")
	for diskId, path := range m.mapDisk {
		h1 := key[:2]
		h2 := key[2:4]

		dir := filepath.Join(path, bucket, h1, h2)
		os.MkdirAll(dir, os.ModePerm)

		// 剩余空间小于500M时候限制上传和同步, *** 暂时去掉大小判断
		freeDisk := DiskUsage(dir).Free
		logger.LogD().Interface("dir", dir).Interface("freedisk", freeDisk).Msg("Get free disk")
		if freeDisk < uint64(env.RESERVED_SPACE+size) {
			err = ErrEnoughSpace
			continue
		}

		return diskId, dir, nil
	}

	return 0, "", err
}

func (m *multiDisk) Put(bucket string, key string, r io.Reader, size int64) error {
	if len(bucket) < 1 || len(key) < 4 {
		logger.LogE().Msg("put file:param error")
		return fmt.Errorf("key(%v) error", key)
	}

	// 如果 betag 已经存在记录，则不需要再次存储。
	if exists, err := m.IsExist(bucket, key); err == nil && exists {
		return nil
	}

	volId, dir, err := m.GenPath(bucket, key, size)
	if err != nil {
		logger.LogE().Msg(fmt.Sprintf("gen dir error:%v", err))
		return err
	}

	filepath := filepath.Join(dir, key)
	if w, err := os.OpenFile(filepath+".tmp", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm); err != nil {
		if os.IsExist(err) {
			fmt.Println(filepath+".tmp", err)
			return nil //已经存在，直接认为成功
		} else {
			logger.LogE().Msg(fmt.Sprintf("open file:%v", err))
			return fmt.Errorf("open:%v", err)
		}
	} else {
		_, err = io.Copy(w, r)
		if err == nil {
			w.Sync() //刷盘，防止断电
			w.Close()
			if err := os.Rename(filepath+".tmp", filepath); err != nil {
				return err
			}

			//加到索引里面
			if err := m.indexer.Add(key, volId); err != nil {
				return err
			}

			//logger.LogI().Interface("pushMsg", attrs).Msg("push uploadMsg to redis")
			//PushMsg(attrs, "put")
			return nil
		} else {
			//删除未能成功上传的文件
			w.Close()
			os.Remove(filepath)
			logger.LogE().Err(fmt.Errorf("copy:%v", err)).Msg("io.Copy  failed")
			return err
		}
	}
}

func (m *multiDisk) Del(bucket string, key string) error {

	fpath, err := m.getFilePath(bucket, key)
	if err != nil {
		return err
	}

	_, err = m.indexer.Delete(key)
	if err != nil {
		return err
	}

	err = os.Remove(fpath)
	logger.LogI().Msg(fmt.Sprintf("push deleteMsg to redis,key:%v", key))
	PushMsg(map[string]interface{}{"key": key, "bucket": bucket}, "delete")
	if err != nil && os.IsNotExist(err) {
		logger.LogE().Msg(fmt.Sprintf("Remove files err:%v", err))
		return nil
	}
	return err
}

func (m *multiDisk) GetFileAbsPath(bucket string, key string) (string, error) {

	filepath, err := m.getFilePath(bucket, key)
	if err != nil {
		return "", err
	}
	return filepath, nil
}

func (m *multiDisk) GetMultipartPath() (int, string) {
	var diskId int
	var path string
	for k, v := range m.mapDisk {
		if diskId < k {
			diskId = k
			path = v
		}
	}
	return diskId, filepath.Join(path, "multipart")
}

func (m *multiDisk) GetPath(bucket string, key string) (string, error) {
	fpath, err := m.getFilePath(bucket, key)
	if err != nil {
		return "", err
	}
	dir, _ := filepath.Split(fpath)
	return dir, nil
}

func (m *multiDisk) GetDiskPathByBEtag(key string) (string, error) {
	diskId, err := m.indexer.Get(key)
	if err != nil {
		return "", err
	}
	diskPath, err := m.GetDiskPath(diskId)
	if err != nil {
		return "", err
	}
	return diskPath, nil
}

func (m *multiDisk) GetRelativePath(bucket string, key string) (string, error) {
	fpath, err := m.getFilePath(bucket, key)
	if err != nil {
		return "", err
	}

	return fpath[len(env.DATA_PATH):], nil
}

func (m *multiDisk) zip(bucket string, uuid string, zw *zip.Writer) error {
	fileInfo, err := dbutils.GetInfoByUuid(uuid)
	if err != nil {
		return err
	}

	if fileInfo.IsDir {
		fileCount := dbutils.CountFileInFolder(fileInfo.Id)
		logger.LogD().Msg(fmt.Sprintf("file in  %s dir count: %d", fileInfo.Name, fileCount))
		if fileCount == 0 {
			logger.LogD().Msg(fmt.Sprintf("blank dir: %s", fileInfo.Name))
			//subDir := strings.TrimLeft(fileInfo.Path, fileInfo.Path+fileInfo.Name)
			subDirPath := filepath.Join("/data/" + fileInfo.Name)
			if _, err := os.Stat(subDirPath); err != nil {
				err := os.Mkdir(subDirPath, os.ModePerm)
				if err != nil {
					logger.LogE().Err(err).Msg("mkdir error")
				}
			}
			subFi, _ := os.Stat(subDirPath)
			header, err := zip.FileInfoHeader(subFi)
			header.Name = fileInfo.Name + "/"
			if err != nil {
				//fmt.Println("error is:"+err.Error())
				return err
			}
			_, err = zw.CreateHeader(header)
			if err != nil {
				//fmt.Println("create error is:"+err.Error())
				return err
			}
			os.Remove(subDirPath)

		}
		if subFileIds, _ := dbutils.GetAllFileInFolder(fileInfo.UserId, uuid); subFileIds != nil {
			for _, subId := range subFileIds {
				subInfo, _ := dbutils.GetInfoByUuid(subId)
				// 空文件夹处理
				if subInfo.IsDir {
					fileCount := dbutils.CountFileInFolder(subInfo.Id)
					logger.LogD().Msg(fmt.Sprintf("file in  %s dir count: %d", subInfo.Name, fileCount))
					if fileCount == 0 {
						logger.LogD().Msg(fmt.Sprintf("blank dir: %s", subInfo.Name))
						subDir := strings.TrimLeft(subInfo.Path, fileInfo.Path+fileInfo.Name)
						subDirPath := filepath.Join("/data/" + subInfo.Name)
						if _, err := os.Stat(subDirPath); err != nil {
							err := os.Mkdir(subDirPath, os.ModePerm)
							if err != nil {
								logger.LogE().Err(err).Msg("mkdir error")
							}
						}
						subFi, _ := os.Stat(subDirPath)
						header, err := zip.FileInfoHeader(subFi)
						header.Name = fileInfo.Name + "/" + subDir + subFi.Name() + "/"
						if err != nil {
							//fmt.Println("error is:"+err.Error())
							return err
						}
						_, err = zw.CreateHeader(header)
						if err != nil {
							//fmt.Println("create error is:"+err.Error())
							return err
						}
						os.Remove(subDirPath)

					}

				}
				subFilePath, err := m.getFilePath(bucket, subInfo.BETag)
				if err != nil {
					return err
				}
				//subDir := strings.Split(subInfo.Path,"/")

				subDir := strings.TrimPrefix(subInfo.Path, fileInfo.Path+fileInfo.Name+"/")

				//在压缩包中创建一个文件
				if !subInfo.IsDir {
					logger.LogD().Msg(fmt.Sprintf("开始压缩：%s", fileInfo.Path+fileInfo.Name+"/"+subDir+subInfo.Name))
					subfile, _ := os.Open(subFilePath)
					subFileOnZip, err := zw.Create(fileInfo.Name + "/" + subDir + subInfo.Name)
					if err != nil {
						logger.LogE().Err(err).Msg(fmt.Sprintf("在压缩包中创建文件失败:%s", fileInfo.Name+"/"+subDir+subInfo.Name))
						subfile.Close()
						return err
					}
					_, err = io.Copy(subFileOnZip, subfile)
					if err != nil {
						logger.LogE().Err(err).Msg(fmt.Sprintf("failed to io.copy %s", subInfo.Id))
						return err
					}
				}

			}

		}

	} else {
		filePath, err := m.getFilePath(bucket, fileInfo.BETag)
		if err != nil {
			return err
		}

		//zipWriter.CreateHeader()
		//在压缩包中创建一个文件
		fileOnZip, err := zw.Create(fileInfo.Name)
		if err != nil {
			logger.LogE().Err(err).Msg(fmt.Sprintf("在压缩包中创建文件失败:%s", fileInfo.Name))
			return err
		}
		//读取需要压缩的文件
		subfile, _ := os.Open(filePath)
		defer subfile.Close()
		_, err = io.Copy(fileOnZip, subfile)
		if err != nil {
			return err
		}

	}
	//defer zw.Close()
	return nil
}