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

package tests

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"aofs/repository/dbutils"
	"aofs/services/multipart"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"aofs/internal/proto"
	"testing"

	"github.com/stretchr/testify/assert"
)

const _2M int64 = 2 * 1024 * 1024
const _4M int64 = _2M * 2
const _8M int64 = _4M * 2
const _16M int64 = _8M * 2

func init() {
	rand.Seed(time.Now().UnixNano())
}

func testMultipartAll(t *testing.T) {
	t.Run("testMultipartLess4M", testMultipartLess4M)

	t.Run("testMultipartConflict", testMultipartConflict)
	//t.Run("testMultipartConflictOfPowerOff", testMultipartConflictOfPowerOff)
	t.Run("testMultipartSpeed", testMultipartSpeed)

	t.Run("testMultipartUploadExecp", testMultipartUploadExecp)
	go t.Run("testMultipartInterface", testMultipartInterface)
	t.Run("testMultipartInterface", testMultipartInterface)
	t.Run("testMultipartConflict", testMultipartConflict)
	t.Run("testMultipartLess8K", testMultipartLess8K)
	t.Run("testMultipartAlbum", testMultipartAlbum)
	t.Run("testMultipartSeq", testMultipartSeq)
	t.Run("testHw", testHw)

	t.Run("testMultipartLess4M_8M", testMultipartLess4M_8M)
	t.Run("testMultipartMore8M", testMultipartMore8M)
	//t.Run("testMultipartMore500M", testMultipartMore500M)
	t.Run("testSizeFlag", testSizeFlag)

}

func testSizeFlag(t *testing.T) {
	assert := assert.New(t)
	m := map[int64]uint8{0: 0, 1: 0, 2: 1, 3: 1, 4: 2, 5: 2, 6: 2, 7: 2, 8: 3}
	for k, v := range m {
		assert.True(multipart.GetSizeFlag(k) == v, "k:%d, v:%d", k, v)
	}
}
func testHw(t *testing.T) {
	assert := assert.New(t)
	hw := &multipart.HashWrite{}
	buf := bytes.NewBuffer(nil)
	hw.W = buf
	hw.Hash = multipart.NewHash()
	filedata := make([]byte, _8M)
	rand.Read(filedata)

	io.CopyN(hw, bytes.NewReader(filedata), int64(len(filedata)))
	sha := md5.Sum(filedata)

	shanew := multipart.NewHash()
	shanew.Write(filedata)
	fmt.Println("shanew:", hex.EncodeToString(shanew.Sum(nil)))
	assert.Equal(hex.EncodeToString(sha[:]), hex.EncodeToString(hw.Sum(nil)))

}

func testMultipartLess8K(t *testing.T) {
	size := int64(8 * 1024)
	testMultipartRandFun(size, t)
}
func testMultipartLess4M(t *testing.T) {
	size := rand.Int63()%(_2M) + _2M
	testMultipartRandFun(size, t)
}
func testMultipartLess4M_8M(t *testing.T) {
	size := rand.Int63()%(_4M) + _4M
	testMultipartRandFun(size, t)
}
func testMultipartMore8M(t *testing.T) {
	size := rand.Int63()%(_8M) + _8M
	testMultipartRandFun(size, t)
}
func testMultipartMore500M(t *testing.T) {
	size := rand.Int63()%(_8M) + 500*1024*1024
	testMultipartRandFun(size, t)
}

func calcFile(file string) {
	st, err := os.Stat(file)
	if err != nil {
		log.Println(err)
		return
	}
	FileSize := st.Size()
	filedata, err := os.ReadFile(file)
	if err != nil {
		log.Println(err)
		return
	}

	hash := multipart.NewHash()

	if FileSize <= multipart.HASH_PART_SIZE {
		hash.Write(filedata)
	} else {
		sumall := []byte{}
		for i := int64(0); i < FileSize; i += multipart.HASH_PART_SIZE {
			start := i
			end := start + multipart.HASH_PART_SIZE
			if end >= FileSize {
				end = FileSize
			}
			sum := md5.Sum(filedata[start:end])
			log.Println("hash:", hex.EncodeToString(sum[:]), "start:", start, "end:", end)
			sumall = append(sumall, sum[:]...)
		}

		hash.Write(sumall)
	}

	etagdata := []byte{multipart.GetSizeFlag(int64(len(filedata)))}
	etagdata = append(etagdata, hash.Sum(nil)...)
	bhash := hex.EncodeToString(etagdata)
	log.Println("betag:", bhash, "filesize:", FileSize)
}

func testMultipartRandFun(size int64, t *testing.T) {

	FileSize := size
	assert := assert.New(t)
	filedata := make([]byte, FileSize)
	rand.Read(filedata)

	hash := multipart.NewHash()

	if FileSize <= multipart.HASH_PART_SIZE {
		hash.Write(filedata)
	} else {
		sumall := []byte{}
		for i := int64(0); i < FileSize; i += multipart.HASH_PART_SIZE {
			start := i
			end := start + multipart.HASH_PART_SIZE
			if end >= FileSize {
				end = FileSize
			}
			sum := md5.Sum(filedata[start:end])
			log.Println("hash:", hex.EncodeToString(sum[:]), start, end, FileSize)
			sumall = append(sumall, sum[:]...)
		}

		hash.Write(sumall)
	}

	etagdata := []byte{multipart.GetSizeFlag(int64(len(filedata)))}
	if multipart.GetSizeFlag(int64(len(filedata)))%2 == 0 {
		etagdata = []byte{}
	}
	etagdata = append(etagdata, hash.Sum(nil)...)
	bhash := hex.EncodeToString(etagdata)

	//ioutil.WriteFile(fmt.Sprintf("%s-size-%d", bhash, FileSize), filedata, os.ModePerm)

	uploadId := ""
	var info *proto.FileInfo
	var isFolderId bool

	//创建任务
	{
		info, _ = dbutils.GetInfoByPath(1, "", "/", proto.TrashStatusNormal)
		var req proto.CreateMultipartTaskReq
		var rsp proto.CreateMultipartTaskRsp
		req.FileName = fmt.Sprintf("%d.txt", time.Now().UnixNano())

		req.Mime = "text/plain"
		req.Size = FileSize
		req.BETag = bhash

		if multipart.GetSizeFlag(int64(len(filedata)))%2 == 0 {
			req.FolderId = info.Id //测试folderId
			isFolderId = true
		} else {
			req.FolderPath = fmt.Sprintf("/%d/", FileSize) //测试folderPath
		}

		var succInfo proto.CreateMultipartTaskSuccRsp

		var rspHead proto.Rsp
		rspHead.Body = &rsp
		rsp.SuccInfo = &succInfo

		TPostRsp("/space/v1/api/multipart/create?userId=1", nil, &req, &rspHead, assert)
		assert.Equal(int(proto.CodeOk), int(rspHead.Code), "%v", rspHead.Message)
		assert.True(len(succInfo.UploadId) > 0 && succInfo.PartSize == _4M, "rsp:%+v", rsp)
		uploadId = succInfo.UploadId

	}

	//进行文件切片
	//parts := []proto.Part{}
	parts := map[int64]proto.Part{}
	for i := int64(0); i < FileSize; {
		size := rand.Int63()%_4M + 1
		part := proto.Part{Start: i, End: i + size}
		if part.End >= FileSize {
			part.End = FileSize
			//parts = append(parts, part)
			parts[part.Start] = part
			break
		}
		//parts = append(parts, part)
		parts[part.Start] = part
		i = part.End
	}

	//upload
	for _, part := range parts {
		sum := md5.Sum(filedata[part.Start:part.End])
		md5sum := hex.EncodeToString(sum[:])
		var rspHead proto.Rsp
		TPostRsp(fmt.Sprintf("/space/v1/api/multipart/upload?userId=1&uploadId=%s&start=%d&end=%d&md5sum=%s", uploadId, part.Start, part.End, md5sum), nil, bytes.NewReader(filedata[part.Start:part.End]), &rspHead, assert)
		assert.Equal(proto.CodeOk, rspHead.Code, "%v", rspHead.Message)
	}

	//complete
	fileid := ""
	{
		var req proto.CompleteMultipartTaskReq
		var rsp proto.CompleteMultipartTaskRsp
		req.UploadId = uploadId

		var rspHead proto.Rsp
		rspHead.Body = &rsp

		TPostRsp("/space/v1/api/multipart/complete?userId=1", nil, &req, &rspHead, assert)
		assert.Equal(int(proto.CodeOk), int(rspHead.Code), "rsp:%+v, uploadid:%v", rsp, uploadId)
		fileid = rsp.Id
		if isFolderId {
			assert.Equal(info.Id, rsp.ParentUuid)
		}

	}

	//文件下载, 比较
	{
		fmt.Println("***************")
		data := TGetFile(fmt.Sprintf("/space/v1/api/file/download?uuid=%s&userId=1", fileid), t)
		assert.True(bytes.Compare(data, filedata) == 0, fileid)
	}

	{

		data := TGetFileRange(fmt.Sprintf("/space/v1/api/file/download?uuid=%s&userId=1", fileid), &proto.Part{}, t)
		assert.Equal(1, len(data))
		assert.Equal(filedata[0], data[0], fileid)
	}
	{

		data := TGetFileRange(fmt.Sprintf("/space/v1/api/file/download?uuid=%s&userId=1", fileid), &proto.Part{Start: 2, End: 3}, t)
		assert.Equal(2, len(data))
		assert.True(bytes.Compare(data, filedata[2:4]) == 0, fileid)
	}

	{

		data := TGetFileRange(fmt.Sprintf("/space/v1/api/file/download?uuid=%s&userId=1", fileid), &proto.Part{Start: 0, End: int64(len(filedata) - 1)}, t)
		assert.Equal(len(filedata), len(data))
		assert.True(bytes.Compare(data, filedata) == 0, fileid)
	}
}

func testMultipartUploadExecp(t *testing.T) {
	assert := assert.New(t)
	filedata := make([]byte, _8M)
	rand.Read(filedata)
	sha1data := md5.Sum(filedata[:_4M])
	sha2data := md5.Sum(filedata[_4M:])
	shadata := append(sha1data[:], sha2data[:]...)
	fmt.Println("sha1data:", hex.EncodeToString(sha1data[:]), "sha2data:", hex.EncodeToString(sha2data[:]))
	shaall := md5.Sum(shadata)
	etagdata := []byte{multipart.GetSizeFlag(int64(len(filedata)))}
	etagdata = append(etagdata, shaall[:]...)
	bhash := hex.EncodeToString(etagdata)
	fileName := fmt.Sprintf("%d.txt", time.Now().UnixNano())
	uploadId := ""
	//创建任务
	{
		info, _ := dbutils.GetInfoByPath(1, "", "/", proto.TrashStatusNormal)
		var req proto.CreateMultipartTaskReq
		var rsp proto.CreateMultipartTaskRsp
		req.FileName = fileName

		req.Mime = "text/plain"
		req.Size = _8M
		req.BETag = bhash
		req.FolderId = info.Id

		var succInfo proto.CreateMultipartTaskSuccRsp
		var rspHead proto.Rsp
		rspHead.Body = &rsp
		rsp.SuccInfo = &succInfo

		TPostRsp("/space/v1/api/multipart/create?userId=1", nil, &req, &rspHead, assert)
		assert.True(len(succInfo.UploadId) > 0 && succInfo.PartSize == _4M, "rsp:%+v", rsp)
		uploadId = succInfo.UploadId
	}

	//已完成分片测试
	{
		start := 0
		end := 1024
		sum := md5.Sum(filedata[start:end])
		md5sum := hex.EncodeToString(sum[:])
		var rspHead proto.Rsp
		TPostRsp(fmt.Sprintf("/space/v1/api/multipart/upload?userId=1&uploadId=%s&start=%d&end=%d&md5sum=%s", uploadId, start, end, md5sum), nil, bytes.NewReader(filedata[start:end]), &rspHead, assert)
		assert.Equal(rspHead.Code, proto.CodeOk)
	}
	{
		start := 2048
		end := 4096
		sum := md5.Sum(filedata[start:end])
		md5sum := hex.EncodeToString(sum[:])
		var rspHead proto.Rsp
		TPostRsp(fmt.Sprintf("/space/v1/api/multipart/upload?userId=1&uploadId=%s&start=%d&end=%d&md5sum=%s", uploadId, start, end, md5sum), nil, bytes.NewReader(filedata[start:end]), &rspHead, assert)
		assert.Equal(rspHead.Code, proto.CodeOk)
	}
	//异常
	{
		start := 1023
		end := 1025
		sum := md5.Sum(filedata[start:end])
		md5sum := hex.EncodeToString(sum[:])
		var rspHead proto.Rsp
		TPostRsp(fmt.Sprintf("/space/v1/api/multipart/upload?userId=1&uploadId=%s&start=%d&end=%d&md5sum=%s", uploadId, start, end, md5sum), nil, bytes.NewReader(filedata[start:end]), &rspHead, assert)
		assert.Equal(rspHead.Code, proto.CodeMultipartTaskOverlap)
	}
	{
		start := 1022
		end := 1023
		sum := md5.Sum(filedata[start:end])
		md5sum := hex.EncodeToString(sum[:])
		var rspHead proto.Rsp
		TPostRsp(fmt.Sprintf("/space/v1/api/multipart/upload?userId=1&uploadId=%s&start=%d&end=%d&md5sum=%s", uploadId, start, end, md5sum), nil, bytes.NewReader(filedata[start:end]), &rspHead, assert)
		assert.Equal(rspHead.Code, proto.CodeMultipartRangeUploaded)
	}
	{
		start := 4096
		end := 4096 + 1
		sum := md5.Sum(filedata[start:end])
		md5sum := hex.EncodeToString(sum[:])
		var rspHead proto.Rsp
		TPostRsp(fmt.Sprintf("/space/v1/api/multipart/upload?userId=1&uploadId=%s&start=%d&end=%d&md5sum=%s", uploadId, start, 9*1024*1024, md5sum), nil, bytes.NewReader(filedata[start:end]), &rspHead, assert)
		assert.Equal(rspHead.Code, proto.CodeMultipartTaskRangeErr)
	}

	{
		start := 1
		end := 4096 + 1
		sum := md5.Sum(filedata[start:end])
		md5sum := hex.EncodeToString(sum[:])
		var rspHead proto.Rsp
		TPostRsp(fmt.Sprintf("/space/v1/api/multipart/upload?userId=1&uploadId=%s&start=%d&end=%d&md5sum=%s", uploadId, start, end, md5sum), nil, bytes.NewReader(filedata[start:end]), &rspHead, assert)
		assert.Equal(rspHead.Code, proto.CodeMultipartTaskOverlap)
	}

	//正在上传分片测试, 速度太快，需要手动，用例暂时屏蔽
	// go func() {
	// 	start := 4096
	// 	end := _2M
	// 	sum := md5.Sum(filedata[start:end])
	// 	md5sum := hex.EncodeToString(sum[:])
	// 	var rspHead proto.Rsp
	// 	TPostRsp(fmt.Sprintf("/space/v1/api/multipart/upload?userId=1&uploadId=%s&start=%d&end=%d&md5sum=%s", uploadId, start, end, md5sum), nil, bytes.NewReader(filedata[start:end]), &rspHead, assert)
	// 	//assert.Equal(rspHead.Code, proto.CodeOk)
	// }()
	// {
	// 	time.Sleep(time.Microsecond * 50)
	// 	start := 4 * 1024
	// 	end := _4M
	// 	sum := md5.Sum(filedata[start:end])
	// 	md5sum := hex.EncodeToString(sum[:])
	// 	var rspHead proto.Rsp
	// 	TPostRsp(fmt.Sprintf("/space/v1/api/multipart/upload?userId=1&uploadId=%s&start=%d&end=%d&md5sum=%s", uploadId, start, end, md5sum), nil, bytes.NewReader(filedata[start:end]), &rspHead, assert)
	// 	assert.Equal(int(rspHead.Code), int(proto.CodeMultipartUploadingConflit))
	// }

}

func testMultipartConflictOfPowerOff(t *testing.T) {
	assert := assert.New(t)
	filedata := make([]byte, _8M)
	rand.Read(filedata)
	sha1data := md5.Sum(filedata[:_4M])
	sha2data := md5.Sum(filedata[_4M:])
	shadata := append(sha1data[:], sha2data[:]...)
	fmt.Println("sha1data:", hex.EncodeToString(sha1data[:]), "sha2data:", hex.EncodeToString(sha2data[:]))
	shaall := md5.Sum(shadata)
	etagdata := []byte{multipart.GetSizeFlag(int64(len(filedata)))}
	etagdata = append(etagdata, shaall[:]...)
	bhash := hex.EncodeToString(etagdata)

	fileName := fmt.Sprintf("%d.txt", time.Now().UnixNano())
	//创建任务
	{
		info, err := dbutils.GetInfoByPath(1, "", "/", proto.TrashStatusNormal)
		if err != nil {
			assert.Fail("failed to get info ", "err:%v", err)
		}
		var req proto.CreateMultipartTaskReq
		var rsp proto.CreateMultipartTaskRsp
		req.FileName = fileName

		req.Mime = "text/plain"
		req.Size = _8M
		req.BETag = bhash
		req.FolderId = info.Id

		var succInfo proto.CreateMultipartTaskSuccRsp
		var rspHead proto.Rsp
		rspHead.Body = &rsp
		rsp.SuccInfo = &succInfo

		TPostRsp("/space/v1/api/multipart/create?userId=1", nil, &req, &rspHead, assert)
		assert.True(len(succInfo.UploadId) > 0 && succInfo.PartSize == _4M, "rsp:%+v", rsp)
	}

	log.Println("shutdown box")

	time.Sleep(time.Minute)
	log.Println("goon")

	//创建任务
	{
		info, err := dbutils.GetInfoByPath(1, "", "/", proto.TrashStatusNormal)
		if err != nil {
			assert.Fail("failed to get info ", "err:%v", err)
		}
		var req proto.CreateMultipartTaskReq
		var rsp proto.CreateMultipartTaskRsp
		req.FileName = fmt.Sprintf("%d222.txt", time.Now().UnixNano())

		req.Mime = "text/plain"
		req.Size = _8M
		req.BETag = bhash
		req.FolderId = info.Id

		var conflictInfo proto.CreateMultipartTaskConflictRsp
		var rspHead proto.Rsp
		rspHead.Body = &rsp
		rsp.ConflictInfo = &conflictInfo

		TPostRsp("/space/v1/api/multipart/create?userId=1", nil, &req, &rspHead, assert)
		assert.Equal(int(proto.CREATE_MULTIPART_TASK_CONFLICT), int(rsp.RspType))
		assert.Equal(fileName, conflictInfo.FileName)

	}

}

func testMultipartConflict(t *testing.T) {
	assert := assert.New(t)
	filedata := make([]byte, _8M)
	rand.Read(filedata)
	sha1data := md5.Sum(filedata[:_4M])
	sha2data := md5.Sum(filedata[_4M:])
	shadata := append(sha1data[:], sha2data[:]...)
	fmt.Println("sha1data:", hex.EncodeToString(sha1data[:]), "sha2data:", hex.EncodeToString(sha2data[:]))
	shaall := md5.Sum(shadata)
	etagdata := []byte{multipart.GetSizeFlag(int64(len(filedata)))}
	etagdata = append(etagdata, shaall[:]...)
	bhash := hex.EncodeToString(etagdata)

	fileName := fmt.Sprintf("%d.txt", time.Now().UnixNano())
	//创建任务
	{
		info, _ := dbutils.GetInfoByPath(1, "", "/", proto.TrashStatusNormal)
		var req proto.CreateMultipartTaskReq
		var rsp proto.CreateMultipartTaskRsp
		req.FileName = fileName

		req.Mime = "text/plain"
		req.Size = _8M
		req.BETag = bhash
		req.FolderId = info.Id

		var succInfo proto.CreateMultipartTaskSuccRsp
		var rspHead proto.Rsp
		rspHead.Body = &rsp
		rsp.SuccInfo = &succInfo

		TPostRsp("/space/v1/api/multipart/create?userId=1", nil, &req, &rspHead, assert)
		assert.True(len(succInfo.UploadId) > 0 && succInfo.PartSize == _4M, "rsp:%+v", rsp)
	}

	//创建任务
	{
		info, _ := dbutils.GetInfoByPath(1, "", "/", proto.TrashStatusNormal)
		var req proto.CreateMultipartTaskReq
		var rsp proto.CreateMultipartTaskRsp
		req.FileName = fmt.Sprintf("%d222.txt", time.Now().UnixNano())

		req.Mime = "text/plain"
		req.Size = _8M
		req.BETag = bhash
		req.FolderId = info.Id

		var conflictInfo proto.CreateMultipartTaskConflictRsp
		var rspHead proto.Rsp
		rspHead.Body = &rsp
		rsp.ConflictInfo = &conflictInfo

		TPostRsp("/space/v1/api/multipart/create?userId=1", nil, &req, &rspHead, assert)
		assert.Equal(int(proto.CREATE_MULTIPART_TASK_CONFLICT), int(rsp.RspType))
		assert.Equal(fileName, conflictInfo.FileName)

	}

}

//测试顺序增加
func testMultipartSeq(t *testing.T) {
	assert := assert.New(t)
	filedata := make([]byte, _8M)
	rand.Read(filedata)
	sha1data := md5.Sum(filedata[:_4M])
	sha2data := md5.Sum(filedata[_4M:])
	shadata := append(sha1data[:], sha2data[:]...)
	fmt.Println("sha1data:", hex.EncodeToString(sha1data[:]), "sha2data:", hex.EncodeToString(sha2data[:]))
	shaall := md5.Sum(shadata)
	etagdata := []byte{multipart.GetSizeFlag(int64(len(filedata)))}
	etagdata = append(etagdata, shaall[:]...)
	bhash := hex.EncodeToString(etagdata)

	uploadId := ""

	//创建任务
	{
		info, _ := dbutils.GetInfoByPath(1, "", "/", proto.TrashStatusNormal)
		var req proto.CreateMultipartTaskReq
		var rsp proto.CreateMultipartTaskRsp
		req.FileName = fmt.Sprintf("%d.txt", time.Now().UnixNano())

		req.Mime = "text/plain"
		req.Size = _8M
		req.BETag = bhash
		req.FolderId = info.Id

		var succInfo proto.CreateMultipartTaskSuccRsp
		var rspHead proto.Rsp
		rspHead.Body = &rsp
		rsp.SuccInfo = &succInfo

		TPostRsp("/space/v1/api/multipart/create?userId=1", nil, &req, &rspHead, assert)
		assert.True(len(succInfo.UploadId) > 0 && succInfo.PartSize == _4M, "rsp:%+v", rsp)
		uploadId = succInfo.UploadId
	}

	//进行文件切片
	parts := []proto.Part{{Start: 0, End: multipart.HASH_PART_SIZE / 2},
		{Start: multipart.HASH_PART_SIZE / 2, End: multipart.HASH_PART_SIZE},
		{Start: multipart.HASH_PART_SIZE, End: multipart.HASH_PART_SIZE * 2}}

	//upload
	for _, part := range parts {
		sum := md5.Sum(filedata[part.Start:part.End])
		md5sum := hex.EncodeToString(sum[:])
		var rspHead proto.Rsp
		TPostRsp(fmt.Sprintf("/space/v1/api/multipart/upload?userId=1&uploadId=%s&start=%d&end=%d&md5sum=%s", uploadId, part.Start, part.End, md5sum), nil, bytes.NewReader(filedata[part.Start:part.End]), &rspHead, assert)
	}

	//complete
	fileid := ""
	{
		var req proto.CompleteMultipartTaskReq
		var rsp proto.CompleteMultipartTaskRsp
		req.UploadId = uploadId

		var rspHead proto.Rsp
		rspHead.Body = &rsp

		TPostRsp("/space/v1/api/multipart/complete?userId=1", nil, &req, &rspHead, assert)
		assert.Equal(int(proto.CodeOk), int(rspHead.Code), "rsp:%+v", rsp)
		fileid = rsp.Id
	}

	//文件下载, 比较
	{
		fmt.Println("***************")
		data := TGetFile(fmt.Sprintf("/space/v1/api/file/download?uuid=%s&userId=1", fileid), t)
		assert.True(bytes.Compare(data, filedata) == 0, "")
	}

	//秒传
	{
		info, _ := dbutils.GetInfoByPath(1, "", "/", proto.TrashStatusNormal)
		var req proto.CreateMultipartTaskReq
		var rsp proto.CreateMultipartTaskRsp
		req.FileName = fmt.Sprintf("%d.txt", time.Now().UnixNano())

		req.Mime = "text/plain"
		req.Size = _8M
		req.BETag = bhash
		req.FolderId = info.Id

		var completeInfo proto.CompleteMultipartTaskRsp
		var rspHead proto.Rsp
		rspHead.Body = &rsp
		rsp.CompleteInfo = &completeInfo

		TPostRsp("/space/v1/api/multipart/create?userId=1", nil, &req, &rspHead, assert)
		assert.Equal(int(proto.CREATE_MULTIPART_TASK_COMPLETE), int(rsp.RspType))

		assert.Equal(req.FileName, completeInfo.Name)

	}
}

//测试速度
func testMultipartSpeed(t *testing.T) {
	times := 128
	assert := assert.New(t)
	filedata := make([]byte, 1024*1024*1024)
	rand.Read(filedata)

	shadata := []byte{}
	for i := 0; i < 256; i++ {
		md5hash := md5.Sum(filedata[i*int(_4M) : (i+1)*int(_4M)])
		shadata = append(shadata, md5hash[:]...)
	}

	shaall := md5.Sum(shadata)
	etagdata := []byte{multipart.GetSizeFlag(int64(len(filedata)))}
	etagdata = append(etagdata, shaall[:]...)
	bhash := hex.EncodeToString(etagdata)

	uploadId := ""

	//创建任务
	{
		info, _ := dbutils.GetInfoByPath(1, "", "/", proto.TrashStatusNormal)
		var req proto.CreateMultipartTaskReq
		var rsp proto.CreateMultipartTaskRsp
		req.FileName = fmt.Sprintf("%d.txt", time.Now().UnixNano())

		req.Mime = "text/plain"
		req.Size = 1024 * 1024 * 1024
		req.BETag = bhash
		req.FolderId = info.Id

		var succInfo proto.CreateMultipartTaskSuccRsp
		var rspHead proto.Rsp
		rspHead.Body = &rsp
		rsp.SuccInfo = &succInfo

		TPostRsp("/space/v1/api/multipart/create?userId=1", nil, &req, &rspHead, assert)
		assert.True(len(succInfo.UploadId) > 0 && succInfo.PartSize == _4M, "rsp:%+v", rsp)
		uploadId = succInfo.UploadId
	}

	ts := time.Now()
	//进行文件切片并发上传
	chBlocks := make(chan int, 10)
	var wg sync.WaitGroup
	wg.Add(times)

	f := func(blockId int) {
		BLOCKSIZE := _4M * (256 / int64(times))
		sum := md5.Sum(filedata[blockId*int(BLOCKSIZE) : (blockId+1)*int(BLOCKSIZE)])
		md5sum := hex.EncodeToString(sum[:])
		var rspHead proto.Rsp
		TPostRsp(fmt.Sprintf("/space/v1/api/multipart/upload?userId=1&uploadId=%s&start=%d&end=%d&md5sum=%s", uploadId, blockId*int(BLOCKSIZE), (blockId+1)*int(BLOCKSIZE), md5sum), nil, bytes.NewReader(filedata[(blockId)*int(BLOCKSIZE):(blockId+1)*int(BLOCKSIZE)]), &rspHead, assert)
		wg.Done()
		<-chBlocks
	}

	go func() {
		for i := 0; i < times; i++ {
			chBlocks <- i
			go f(i)
		}
	}()

	wg.Wait()

	speed := float64(1024) / (float64(time.Now().Sub(ts)) / float64(time.Second))
	fmt.Printf("********* speed: %f m/s\n", speed)

	//complete
	{
		var req proto.CompleteMultipartTaskReq
		var rsp proto.CompleteMultipartTaskRsp
		req.UploadId = uploadId

		var rspHead proto.Rsp
		rspHead.Body = &rsp

		TPostRsp("/space/v1/api/multipart/complete?userId=1", nil, &req, &rspHead, assert)
		assert.Equal(int(proto.CodeOk), int(rspHead.Code), "rsp:%+v", rsp)
	}
}

var folderId string

func init() {
	folderId = fmt.Sprintf("/%d/%d/", time.Now().UnixNano(), time.Now().UnixNano())
}
func testMultipartInterface(t *testing.T) {
	// defer func() {
	// 	os.Exit(1)
	// }()
	assert := assert.New(t)
	filedata := make([]byte, _8M)
	rand.Read(filedata)
	sha1data := md5.Sum(filedata[:_4M])
	sha2data := md5.Sum(filedata[_4M:])
	shadata := append(sha1data[:], sha2data[:]...)
	fmt.Println("sha1data:", hex.EncodeToString(sha1data[:]), "sha2data:", hex.EncodeToString(sha2data[:]))
	shaall := md5.Sum(shadata)
	etagdata := []byte{multipart.GetSizeFlag(int64(len(filedata)))}
	etagdata = append(etagdata, shaall[:]...)
	bhash := hex.EncodeToString(etagdata)

	uploadId := ""

	//创建任务
	{

		var req proto.CreateMultipartTaskReq
		var rsp proto.CreateMultipartTaskRsp
		req.FileName = fmt.Sprintf("%d.txt", time.Now().UnixNano())

		req.Mime = "text/plain"
		req.Size = _8M
		req.BETag = bhash
		req.FolderPath = folderId // "/111/" //fmt.Sprintf("/%d/", time.Now().UnixNano())

		var succInfo proto.CreateMultipartTaskSuccRsp
		var rspHead proto.Rsp
		rspHead.Body = &rsp
		rsp.SuccInfo = &succInfo

		TPostRsp("/space/v1/api/multipart/create?userId=1", nil, &req, &rspHead, assert)
		assert.True(len(succInfo.UploadId) > 0 && succInfo.PartSize == _4M, "rsp:%+v", rsp)
		uploadId = succInfo.UploadId
	}

	// list
	{
		var req proto.ListMultipartReq
		var rsp proto.ListMultipartRsp
		req.UploadId = "xxx"

		var rspHead proto.Rsp
		rspHead.Body = &rsp

		TGetRsp(fmt.Sprintf("/space/v1/api/multipart/list?userId=1&uploadId=%s", uploadId), &rspHead, assert)
		assert.Equal(proto.CodeOk, rspHead.Code, "rsp:%+v", rspHead)
		assert.Equal(0, len(rsp.UploadedParts), "rsp:%+v", rsp)
	}

	//upload
	{
		sum := md5.Sum(filedata[:_4M])
		md5sum := hex.EncodeToString(sum[:])
		var rspHead proto.Rsp
		TPostRsp(fmt.Sprintf("/space/v1/api/multipart/upload?userId=1&uploadId=%s&start=0&end=%d&md5sum=%s", uploadId, _4M, md5sum), nil, bytes.NewReader(filedata[:_4M]), &rspHead, assert)
	}

	// list
	{

		var rsp proto.ListMultipartRsp

		var rspHead proto.Rsp
		rspHead.Body = &rsp

		TGetRsp(fmt.Sprintf("/space/v1/api/multipart/list?userId=1&uploadId=%s", uploadId), &rspHead, assert)
		assert.True(len(rsp.UploadedParts) == 1, "rsp:%+v", rsp)
	}

	//upload
	{
		sum := md5.Sum(filedata[_4M:])
		md5sum := hex.EncodeToString(sum[:])
		var rspHead proto.Rsp
		TPostRsp(fmt.Sprintf("/space/v1/api/multipart/upload?userId=1&uploadId=%s&start=%d&end=%d&md5sum=%s", uploadId, _4M, len(filedata), md5sum), nil, bytes.NewReader(filedata[_4M:]), &rspHead, assert)
	}

	// list
	{

		var rsp proto.ListMultipartRsp

		var rspHead proto.Rsp
		rspHead.Body = &rsp

		TGetRsp(fmt.Sprintf("/space/v1/api/multipart/list?userId=1&uploadId=%s", uploadId), &rspHead, assert)
		assert.True(len(rsp.UploadedParts) == 1, "rsp:%+v", rsp)
	}

	//complete
	fileid := ""
	{
		var req proto.CompleteMultipartTaskReq
		var rsp proto.CompleteMultipartTaskRsp
		req.UploadId = uploadId

		var rspHead proto.Rsp
		rspHead.Body = &rsp

		TPostRsp("/space/v1/api/multipart/complete?userId=1", nil, &req, &rspHead, assert)
		assert.Equal(int(proto.CodeOk), int(rspHead.Code), "rsp:%+v", rsp)
		fileid = rsp.Id
	}

	//文件下载, 比较
	{
		fmt.Println("***************")
		data := TGetFile(fmt.Sprintf("/space/v1/api/file/download?uuid=%s&userId=1", fileid), t)
		assert.True(bytes.Compare(data, filedata) == 0, "")
	}

	//delete
	{
		var req proto.DeleteMultipartTaskReq
		req.UploadId = uploadId

		var rspHead proto.Rsp

		TPostRsp("/space/v1/api/multipart/delete?userId=1", nil, &req, &rspHead, assert)
		assert.True(rspHead.Code == proto.CodeOk, "rsp:%+v", rspHead)
	}

	//os.Exit(1)
}

func testMultipartAlbum(t *testing.T) {
	assert := assert.New(t)
	filedata := make([]byte, _8M)
	rand.Read(filedata)
	sha1data := md5.Sum(filedata[:_4M])
	sha2data := md5.Sum(filedata[_4M:])
	shadata := append(sha1data[:], sha2data[:]...)
	fmt.Println("sha1data:", hex.EncodeToString(sha1data[:]), "sha2data:", hex.EncodeToString(sha2data[:]))
	shaall := md5.Sum(shadata)
	etagdata := []byte{multipart.GetSizeFlag(int64(len(filedata)))}
	etagdata = append(etagdata, shaall[:]...)
	bhash := hex.EncodeToString(etagdata)

	uploadId := ""

	//创建任务
	{
		info, _ := dbutils.GetInfoByPath(1, "", "/", proto.TrashStatusNormal)
		var req proto.CreateMultipartTaskReq
		var rsp proto.CreateMultipartTaskRsp
		req.FileName = fmt.Sprintf("%d.txt", time.Now().UnixNano())

		req.Mime = "text/plain"
		req.Size = _8M
		req.BETag = bhash
		req.FolderId = info.Id

		var succInfo proto.CreateMultipartTaskSuccRsp
		var rspHead proto.Rsp
		rspHead.Body = &rsp
		rsp.SuccInfo = &succInfo

		TPostRsp("/space/v1/api/multipart/create?userId=1", nil, &req, &rspHead, assert)
		assert.True(len(succInfo.UploadId) > 0 && succInfo.PartSize == _4M, "rsp:%+v", rsp)
		uploadId = succInfo.UploadId
	}

	//进行文件切片
	parts := []proto.Part{{Start: 0, End: multipart.HASH_PART_SIZE / 2},
		{Start: multipart.HASH_PART_SIZE / 2, End: multipart.HASH_PART_SIZE},
		{Start: multipart.HASH_PART_SIZE, End: multipart.HASH_PART_SIZE * 2}}

	//upload
	for _, part := range parts {
		sum := md5.Sum(filedata[part.Start:part.End])
		md5sum := hex.EncodeToString(sum[:])
		var rspHead proto.Rsp
		TPostRsp(fmt.Sprintf("/space/v1/api/multipart/upload?userId=1&uploadId=%s&start=%d&end=%d&md5sum=%s", uploadId, part.Start, part.End, md5sum), nil, bytes.NewReader(filedata[part.Start:part.End]), &rspHead, assert)
	}

	//complete
	fileid := ""
	{
		var req proto.CompleteMultipartTaskReq
		var rsp proto.CompleteMultipartTaskRsp
		req.UploadId = uploadId

		var rspHead proto.Rsp
		rspHead.Body = &rsp

		TPostRsp("/space/v1/api/multipart/complete?userId=1", nil, &req, &rspHead, assert)
		assert.Equal(int(proto.CodeOk), int(rspHead.Code), "rsp:%+v", rsp)
		fileid = rsp.Id
	}

	//文件下载, 比较
	{
		fmt.Println("***************")
		data := TGetFile(fmt.Sprintf("/space/v1/api/file/download?uuid=%s&userId=1", fileid), t)
		assert.True(bytes.Compare(data, filedata) == 0, "")
	}

	//普通秒传
	sameFilename := fmt.Sprintf("%d.txt", time.Now().UnixNano())
	{
		info, _ := dbutils.GetInfoByPath(1, "", "/", proto.TrashStatusNormal)
		var req proto.CreateMultipartTaskReq
		var rsp proto.CreateMultipartTaskRsp
		req.FileName = sameFilename

		req.Mime = "text/plain"
		req.Size = _8M
		req.BETag = bhash
		req.FolderId = info.Id

		var completeInfo proto.CompleteMultipartTaskRsp
		var rspHead proto.Rsp
		rspHead.Body = &rsp
		rsp.CompleteInfo = &completeInfo

		TPostRsp("/space/v1/api/multipart/create?userId=1", nil, &req, &rspHead, assert)
		assert.Equal(int(proto.CREATE_MULTIPART_TASK_COMPLETE), int(rsp.RspType))

		assert.Equal(req.FileName, completeInfo.Name)
	}

	{
		info, _ := dbutils.GetInfoByPath(1, "", "/", proto.TrashStatusNormal)
		var req proto.CreateMultipartTaskReq
		var rsp proto.CreateMultipartTaskRsp
		req.FileName = sameFilename

		req.Mime = "text/plain"
		req.Size = _8M
		req.BETag = bhash
		req.FolderId = info.Id

		var completeInfo proto.CompleteMultipartTaskRsp
		var rspHead proto.Rsp
		rspHead.Body = &rsp
		rsp.CompleteInfo = &completeInfo

		TPostRsp("/space/v1/api/multipart/create?userId=1", nil, &req, &rspHead, assert)
		assert.Equal(int(proto.CREATE_MULTIPART_TASK_COMPLETE), int(rsp.RspType))

		//直接覆盖？
		//assert.NotEqual(req.FileName, completeInfo.Name)
	}

	//秒传
	{
		info, _ := dbutils.GetInfoByPath(1, "", "/", proto.TrashStatusNormal)
		var req proto.CreateMultipartTaskReq
		var rsp proto.CreateMultipartTaskRsp
		req.FileName = sameFilename

		req.Mime = "text/plain"
		req.Size = _8M
		req.BETag = bhash
		req.FolderId = info.Id
	

		var completeInfo proto.CompleteMultipartTaskRsp
		var rspHead proto.Rsp
		rspHead.Body = &rsp
		rsp.CompleteInfo = &completeInfo

		TPostRsp("/space/v1/api/multipart/create?userId=1", nil, &req, &rspHead, assert)
		assert.Equal(int(proto.CREATE_MULTIPART_TASK_COMPLETE), int(rsp.RspType))

		assert.Equal(req.FileName, completeInfo.Name)
	}
}

//func TestGetCharset(t *testing.T) {
//	assert := assert.New(t)
//	param := proto.CreateMultipartTaskReq{
//		BETag: "0385fb99a9b4be08806ce2f5da0574fa8d",
//	}
//
//	//mime := "text/html"
//
//	name := GetCharset(param.BETag)
//
//	assert.Equal("GB2312",name)
//
//}
