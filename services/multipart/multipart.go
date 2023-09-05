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

package multipart

import (
	"aofs/internal/bpctx"
	"aofs/internal/env"
	"aofs/internal/log4bp"
	"aofs/internal/proto"
	"aofs/repository/dbutils"
	"aofs/repository/storage"
	"crypto/md5"

	"encoding/hex"
	"encoding/json"

	"aofs/internal/utils"
	"fmt"
	"hash"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var stor storage.MultiDiskStorager

var logger = log4bp.New("", gin.Mode())

//此文件完成分片上传接口实现

const HASH_PART_SIZE = 4 * 1024 * 1024
const HASH_SUM_SIZE = 16

func NewHash() hash.Hash {
	return md5.New()
}

func GetSizeFlag(size int64) uint8 {
	n := uint8(0)
	for size/2 >= 1 {
		n++
		size /= 2
	}
	return n
}

type MultipartTask struct {
	UploadId       string                         //上传任务id
	DiskId         int                            //存储的磁盘id
	MPDataPath     string                         //数据存储路径, 包括 hash 和 data 文件，分片记录文件存在主目录下。
	Param          proto.CreateMultipartTaskParam //创建任务时的参数
	HashPartSize   int64                          // hash分片大小
	UploadedParts  []proto.Part                   //已上传的分片
	UploadingParts []proto.Part                   //正在上传的分片
	last           int64                          //最后的活跃时间
	mutex          sync.Mutex
	betagPath      string //betag 对应的文件存储路径
	diskPath       string
}

func (task *MultipartTask) UpdateLast() {
	task.last = time.Now().Unix()
}
func (task *MultipartTask) timeout() bool {
	return (time.Now().Unix() - task.last) > 30*60 //30分钟超时
}

func (task *MultipartTask) GetTaskInfo() proto.MultipartTaskStatusInfo {
	task.mutex.Lock()
	defer task.mutex.Unlock()
	var info proto.MultipartTaskStatusInfo
	info.UploadId = task.UploadId
	info.CreateMultipartTaskParam = task.Param
	info.UploadedParts = task.UploadedParts
	info.UploadingParts = task.UploadingParts
	return info
}

func (task *MultipartTask) GetBHashPartSize() int64 {
	return 4 * 1024 * 1024
}

func (task *MultipartTask) Complete() error {
	task.mutex.Lock()
	defer task.mutex.Unlock()
	if len(task.UploadingParts) != 0 {
		return fmt.Errorf("uploading, uploadingParts:%v", task.UploadingParts)
	}

	if len(task.UploadedParts) != 1 {
		return fmt.Errorf("parts is not as expected, uploadedParts:%v", task.UploadedParts)
	}

	if task.UploadedParts[0].Start != 0 || task.UploadedParts[0].End != task.Param.Size {
		return fmt.Errorf("not finish. uploadedParts:%v, task.param.size=%v", task.UploadedParts, task.Param.Size)
	}

	//校验
	file, err := os.Open(filepath.Join(task.MPDataPath, task.UploadId+".hash"))
	if err != nil {
		return fmt.Errorf("failed to open hash file")
	}
	var shasum []byte
	if task.Param.Size <= HASH_PART_SIZE {
		shasum, err = ioutil.ReadAll(file)
		file.Close()

		if len(shasum) != HASH_SUM_SIZE || err != nil {
			return fmt.Errorf("failed to read hash file")
		}

	} else {
		sha := NewHash()
		io.Copy(sha, file)

		file.Close()
		shasum = sha.Sum(nil)

	}
	shasum = append([]byte{GetSizeFlag(task.Param.Size)}, shasum...)

	if hex.EncodeToString(shasum) != task.Param.BETag {
		return fmt.Errorf("hash err, actual:%v, expect:%v", hex.EncodeToString(shasum), task.Param.BETag)
	}

	//校验通过
	if fpath, err := stor.MoveFile(filepath.Join(task.MPDataPath, task.UploadId+".data"), task.DiskId, env.NORMAL_BUCKET, task.UploadId); err != nil {
		return err
	} else {
		task.betagPath = fpath
		if err := os.Remove(filepath.Join(task.MPDataPath, task.UploadId+".hash")); err != nil {
			log.Println("failed to remove hash:", err)
		}
		os.Remove(filepath.Join(getMPMetaPath(), task.UploadId+".mp"))
		return nil
	}

}

type HashWrite struct {
	hash.Hash
	W io.Writer
}

func (hw *HashWrite) Write(b []byte) (int, error) {
	n, err := hw.W.Write(b)
	if n > 0 && hw.Hash != nil {
		hw.Hash.Write(b)
	}
	return n, err
}

func (task *MultipartTask) checkUploadedOverlap(start, end int64) (bool, error) {
	for _, part := range task.UploadedParts {
		if start >= part.Start && end <= part.End {
			return true, nil
		} else if start < part.End && end > part.Start {
			return true, fmt.Errorf("overlap to uploaded")
		}
	}
	return false, nil
}

func (task *MultipartTask) checkUploadingOverlap(start, end int64) (bool, error) {
	for _, part := range task.UploadingParts {
		if start >= part.Start && end <= part.End {
			return true, nil
		} else if start < part.End && end > part.Start {
			return true, fmt.Errorf("overlap to uploading")
		}
	}
	return false, nil
}

func (task *MultipartTask) removeFromUploading(start int64, lock bool) {
	if lock {
		task.mutex.Lock()
		defer task.mutex.Unlock()
	}

	for i, part := range task.UploadingParts {
		if part.Start == start {
			task.UploadingParts = append(task.UploadingParts[:i], task.UploadingParts[i+1:]...)
			break
		}
	}
}

//合并当前完成的分段，返回需要计算hash的分段, 支持随机分片长度。
func (task *MultipartTask) mergeToUploaded(start, end int64, lock bool) []int64 {
	ret := []int64{}
	if lock {
		task.mutex.Lock()
		defer task.mutex.Unlock()
	}
	task.UploadedParts = append(task.UploadedParts, proto.Part{Start: start, End: end})
	sort.Slice(task.UploadedParts, func(i, j int) bool {
		return task.UploadedParts[i].Start < task.UploadedParts[j].Start
	})

	//合并
	subparts := []proto.Part{}
	cur := task.UploadedParts[0]
	for _, part := range task.UploadedParts[1:] {
		if part.Start == cur.End {
			//找出需要计算的
			if cur.End%HASH_PART_SIZE != 0 &&
				(part.End/HASH_PART_SIZE > cur.End/HASH_PART_SIZE || part.End == task.Param.Size) {
				ret = append(ret, cur.End/HASH_PART_SIZE)
			}
			cur.End = part.End
		} else {
			subparts = append(subparts, cur)
			cur = part
		}
	}
	task.UploadedParts = append(subparts, cur)

	return ret
}

func (task *MultipartTask) splitPart(start, end int64) []proto.Part {
	parts := []proto.Part{}
	for {
		part := proto.Part{Start: start, End: start + HASH_PART_SIZE}
		part.End -= part.End % HASH_PART_SIZE
		if part.End >= end {
			part.End = end
			parts = append(parts, part)
			break
		} else {
			parts = append(parts, part)
			start = part.End
		}
	}
	return parts
}

//写文件，如果本次分片包含完整的hash计算端，则在内存完成计算；否则不计算，留待后续合并再计算
func (task *MultipartTask) writeDatafile(start, end int64, r io.Reader) (map[int64][]byte, error) {

	w, err := os.OpenFile(filepath.Join(task.MPDataPath, task.UploadId+".data"), os.O_WRONLY, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to open file:%v", err)
	}
	defer w.Close()
	defer w.Sync()
	if pos, err := w.Seek(start, io.SeekStart); err != nil || pos != start {
		return nil, fmt.Errorf("failed to seek.%v", err)
	} else {
		//进行分段计算，找出可以内存计算hash的段
		hashMaps := map[int64][]byte{}
		parts := task.splitPart(start, end)
		for _, part := range parts {
			hw := &HashWrite{W: w}
			if part.Start%HASH_PART_SIZE == 0 && (part.End%HASH_PART_SIZE == 0 || part.End == task.Param.Size) {
				hw.Hash = NewHash()
			}

			if _, err := io.CopyN(hw, r, part.End-part.Start); err != nil {
				return nil, fmt.Errorf("failed to copy.%v", err)
			}

			if hw.Hash != nil {
				hashMaps[part.Start/HASH_PART_SIZE] = hw.Sum(nil)
			}
		}

		return hashMaps, nil
	}
}

func (task *MultipartTask) getPartHash(pos int64) ([]byte, error) {

	datafile, err := os.OpenFile(filepath.Join(task.MPDataPath, task.UploadId+".data"), os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to open file:%v", err)
	}
	defer datafile.Close()
	start := pos * HASH_PART_SIZE
	if pos, err := datafile.Seek(start, io.SeekStart); err != nil || pos != start {
		return nil, err
	}
	end := start + HASH_PART_SIZE
	if end > task.Param.Size {
		end = task.Param.Size
	}

	pr := io.LimitReader(datafile, end-start)

	sha := NewHash()
	if n, err := io.Copy(sha, pr); err != nil {
		return nil, err
	} else if n != (end - start) {
		return nil, fmt.Errorf("sha size err")
	}

	return sha.Sum(nil), nil
}

func (task *MultipartTask) writeMpfile() error {
	task.mutex.Lock()
	defer task.mutex.Unlock()
	mpfile, err := os.OpenFile(filepath.Join(getMPMetaPath(), task.UploadId+".mp"), os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to write mp file:%v", err)
	}
	defer mpfile.Close()
	defer mpfile.Sync()

	d, _ := json.Marshal(task)
	if n, err := mpfile.Write(d); n != len(d) || err != nil {
		return fmt.Errorf("failed to write mp:%v", err)
	}

	return nil

}

//写入第几段hash， 如果hash值为空，则从文件读取
func (task *MultipartTask) writeHashfile(posHash map[int64][]byte) error {
	//写hash文件
	hashfile, err := os.OpenFile(filepath.Join(task.MPDataPath, task.UploadId+".hash"), os.O_WRONLY, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to write hash file:%v", err)
	}
	defer hashfile.Close()
	defer hashfile.Sync()

	for pos, hashData := range posHash {
		seekpos, err := hashfile.Seek(pos*HASH_SUM_SIZE, io.SeekStart)
		if err != nil || seekpos != pos*HASH_SUM_SIZE {
			return fmt.Errorf("failed to seek hash file")
		}
		if hashData == nil {
			//从文件扒出来做hash
			if hashData, err = task.getPartHash(pos); err != nil {
				return err
			}
		}

		if n, err := hashfile.Write(hashData); err != nil || n != HASH_SUM_SIZE {
			return fmt.Errorf("failed to write hash file:n=%d, err=%v", n, err)
		}
	}

	return nil
}

func (task *MultipartTask) Upload(start, end int64, r io.Reader) (proto.CodeType, error) {
	//检测文件是否出错
	{
		task.mutex.Lock()
		task.last = time.Now().Unix()
		if start < 0 || end > task.Param.Size || start >= end {
			task.mutex.Unlock()
			return proto.CodeMultipartTaskRangeErr, fmt.Errorf("pos invalid")
		}

		//检查重叠
		if isOverlap, err := task.checkUploadedOverlap(start, end); isOverlap && err == nil {
			task.mutex.Unlock()
			return proto.CodeMultipartRangeUploaded, fmt.Errorf("range uploaded")
		} else if isOverlap {
			task.mutex.Unlock()
			return proto.CodeMultipartTaskOverlap, fmt.Errorf("uploaded overlap")
		}

		if isOverlap, _ := task.checkUploadingOverlap(start, end); isOverlap {
			task.mutex.Unlock()
			return proto.CodeMultipartUploadingConflit, fmt.Errorf("uploading conflit")
		}

		task.UploadingParts = append(task.UploadingParts, proto.Part{Start: start, End: end})
		sort.Slice(task.UploadingParts, func(i, j int) bool {
			return task.UploadingParts[i].Start < task.UploadingParts[j].Start
		})
		task.mutex.Unlock()
	}

	//如果出错则删除正在上传中的记录
	defer task.removeFromUploading(start, true)

	//正式上传
	hashData, err := task.writeDatafile(start, end, r)
	if err != nil {
		return proto.CodeFailedToSaveFile, err
	} else if len(hashData) > 0 {
		//写入hash
		err = task.writeHashfile(hashData)
		if err != nil {
			return proto.CodeFailedToSaveFile, err
		}
	}

	//copy完成, 保存hash文件
	parts := task.mergeToUploaded(start, end, true)
	for _, pos := range parts {
		err = task.writeHashfile(map[int64][]byte{pos: nil})
		if err != nil {
			return proto.CodeFailedToSaveFile, err
		}
	}
	//保存记录文件

	if err := task.writeMpfile(); err != nil {
		return proto.CodeFailedToSaveFile, err
	}

	return proto.CodeOk, nil
}

type multipartTaskMng struct {
	mutex   sync.Mutex
	mapTask map[string]*MultipartTask
}

func (mtm *multipartTaskMng) lru() {
	lastDay := 0
	for {

		mtm.mutex.Lock()
		for k, v := range mtm.mapTask {
			if v.timeout() {
				delete(mtm.mapTask, k)
				break
			}
		}
		mtm.mutex.Unlock()
		if time.Now().Day() != lastDay {
			filepath.Walk(getMPMetaPath(), mtm.clearExpiredTask)
			lastDay = time.Now().Day()
		}

		time.Sleep(time.Minute * 5)
	}
}

func (mtm *multipartTaskMng) clearExpiredTask(path string, info fs.FileInfo, err error) error {
	if !info.IsDir() && filepath.Ext(info.Name()) == ".mp" && int(time.Now().Unix()-info.ModTime().Unix()) > env.MULTIPART_TASK_LRU_SECOND {
		etag := info.Name()[:len(info.Name())-3]
		mtm.DeleteTask(etag)
		log.Println("delete task files:", etag)
	}
	return nil
}

func (mtm *multipartTaskMng) GetTask(uploadId string) (*MultipartTask, error) {
	mtm.mutex.Lock()
	defer mtm.mutex.Unlock()

	//内存中查找
	if task, ok := mtm.mapTask[uploadId]; ok {
		return task, nil
	}

	path := filepath.Join(getMPMetaPath(), uploadId+".mp")
	task := &MultipartTask{}
	if data, err := ioutil.ReadFile(path); err == nil {
		json.Unmarshal(data, task)
		task.UploadingParts = nil
		task.UpdateLast()
		mtm.mapTask[uploadId] = task
		//此处需要判断是否是同一个文件,请求参数是否一致
		return task, nil
	} else {
		logger.LogE().Msg("not found upload task")
		return nil, fmt.Errorf("not found task:%v", err)
	}

}

func getMPMetaPath() string {
	return filepath.Join(env.DATA_PATH, "multipart-meta")
}

func (mtm *multipartTaskMng) GenTask(param proto.CreateMultipartTaskReq) (*MultipartTask, error) {
	mtm.mutex.Lock()
	defer mtm.mutex.Unlock()

	//内存中查找
	if task, ok := mtm.mapTask[param.BETag]; ok {
		return task, os.ErrExist
	}

	//从文件加载
	dir := getMPMetaPath()
	path := filepath.Join(dir, param.BETag+".mp")
	task := &MultipartTask{}

	if data, err := ioutil.ReadFile(path); err == nil && len(data) != 0 {
		if err := json.Unmarshal(data, task); err != nil {
			os.Remove(path)
			return nil, fmt.Errorf("failed to marshal mp file:%v", err)
		}
		task.UploadingParts = nil
		task.UpdateLast()
		mtm.mapTask[param.BETag] = task
		//此处需要判断是否是同一个文件,请求参数是否一致
		return task, os.ErrExist

	} else {
		task.Param = param
		task.UploadId = param.BETag
		task.DiskId, _, err = stor.GenPath(env.NORMAL_BUCKET, task.UploadId, task.Param.Size)
		if err != nil {
			return task, err
		}
		task.MPDataPath, _ = stor.GetDiskMPPath(task.DiskId)
		task.diskPath, _ = stor.GetDiskPath(task.DiskId)

		data, _ = json.Marshal(task)

		ioutil.WriteFile(path, data, os.ModePerm)

		//创建数据文件和hash文件
		ioutil.WriteFile(filepath.Join(task.MPDataPath, task.UploadId+".hash"), []byte{}, os.ModePerm)
		ioutil.WriteFile(filepath.Join(task.MPDataPath, task.UploadId+".data"), []byte{}, os.ModePerm)
		mtm.mapTask[param.BETag] = task
		return task, nil
	}

}

func (mtm *multipartTaskMng) DeleteTask(uploadId string) error {
	mtm.mutex.Lock()
	defer mtm.mutex.Unlock()

	if task, ok := mtm.mapTask[uploadId]; ok && len(task.UploadingParts) > 0 {
		logger.LogW().Msg("uploading")
		return fmt.Errorf("uploading")
	} else {
		delete(mtm.mapTask, uploadId)

		os.Remove(filepath.Join(getMPMetaPath(), uploadId+".mp"))
		if task != nil {
			os.Remove(filepath.Join(task.MPDataPath, uploadId+".data"))
			os.Remove(filepath.Join(task.MPDataPath, uploadId+".hash"))
		}

		return nil
	}

}

func (mtm *multipartTaskMng) RemoveTask(uploadId string) {
	mtm.mutex.Lock()
	defer mtm.mutex.Unlock()
	delete(mtm.mapTask, uploadId)
}

var Taskmgr multipartTaskMng

func Init() {
	stor = storage.GetStor()
	Taskmgr.mapTask = make(map[string]*MultipartTask)
	go Taskmgr.lru()
}

func InsertIndex(ctx *bpctx.Context, param proto.CreateMultipartTaskReq, isUploadData bool, task *MultipartTask) (proto.FileInfo, error) {

	var extJson []byte
	if utils.GetMimeTypeByFilename(param.FileName) == "text/plain" || utils.GetMimeTypeByFilename(param.FileName) == "text/html" {
		ext := proto.FileInfoExt{
			Charset: storage.GetCharset(param.BETag),
		}
		extJson, _ = json.Marshal(ext)
	}
	//任务完成上传，创建索引
	fileinfo := proto.FileInfo{
		FileInfoPub: proto.FileInfoPub{
			Id:            uuid.New().String(),
			ParentUuid:    param.FolderId,
			IsDir:         false,
			Name:          param.FileName,
			Path:          param.FolderPath,
			BETag:         param.BETag,
			CreateTime:    time.Now().UnixNano() / 1e6,
			ModifyTime:    time.Now().UnixNano() / 1e6,
			OperationTime: time.Now().UnixNano() / 1e6,
			Size:          param.Size,
			FileCount:     0,
			Category:      utils.ParseCategoryByFilename(param.FileName),
			Mime:          utils.GetMimeTypeByFilename(param.FileName),
		},

		UserId:      ctx.GetUserId(),
		BucketName:  env.NORMAL_BUCKET,
		FileInfoExt: extJson,
	}
	if param.CreateTime > 0 {
		fileinfo.CreateTime = param.CreateTime
	}
	if param.ModifyTime > 0 {
		fileinfo.ModifyTime = param.ModifyTime
	}

	// 判断是否有同名文件
	fi, err := dbutils.GetInfoByPath(ctx.GetUserId(), param.FolderPath, param.FileName, proto.TrashStatusNormal)
	if err == nil {
		//betag相同文件处理
		if param.BETag == fi.BETag {
			//直接覆盖
			if err := dbutils.UpdateOperationTime(fi.Id); err != nil {
				return *fi, err
			} else {
				return *fi, nil
			}
		} else {
			//改名上传
			newName, err := dbutils.GenIncNameByPath(ctx.GetUserId(), param.FolderPath, param.FileName, proto.TrashStatusNormal)
			if err != nil {
				return fileinfo, err
			}
			fileinfo.Name = newName
		}
	}

	err = dbutils.AddFileV2(fileinfo, fileinfo.ParentUuid)
	if err == nil && isUploadData {
		attrs := map[string]interface{}{"key": fileinfo.BETag,
			"betagPath":   task.betagPath,
			"diskPath":    task.diskPath,
			"size":        fileinfo.Size,
			"name":        fileinfo.Name,
			"bucket":      fileinfo.BucketName,
			"contentType": fileinfo.Mime}
		storage.PushMsg(attrs, "put")
	}

	return fileinfo, err

}
