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
	"aofs/internal/proto"
	"aofs/repository/dbutils"
	"aofs/routers/api"
	"aofs/services/async"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func testFilesAll(t *testing.T) {
	//t.Run("testVodSymlink", testVodSymlink)
	//return

	t.Run("testFilesCopy", testFilesCopy)
	t.Run("testFilesList", testFilesList)
	t.Run("testFilesRename", testFilesRename)
	t.Run("testFolderRename", testFolderRename)
	t.Run("testFilesMove", testFilesMove)
}

// func testVodSymlink(t *testing.T) {
// 	assert := assert.New(t)
// 	var req proto.SoftlinkReq
// 	var linkRsp proto.SoftlinkRsp
// 	var rsp proto.Rsp
// 	rsp.Body = &linkRsp

// 	fi, err := dbutils.GetInfoByPath(1, "/", "说明.pdf", proto.TrashStatusNormal)
// 	assert.Equal(nil, err, "文档目录不存在:%v", err)

// 	req.Id = fi.Id

// 	{
// 		TPostRsp(fmt.Sprintf("/space/v1/api/file/softlink?userId=1"), nil, &req, &rsp, assert)
// 		assert.Equal(int(proto.CodeOk), int(rsp.Code))
// 	}

// }

func testFilesCopy(t *testing.T) {
	assert := assert.New(t)
	var req proto.CopyFileReq
	var copyRsp proto.CopyRsp
	var rsp proto.Rsp
	rsp.Body = &copyRsp

	fi, err := dbutils.GetInfoByPath(1, "/", "文档", proto.TrashStatusNormal)
	assert.Equal(nil, err, "文档目录不存在:%v", err)

	req.DestId = fi.Id

	{
		TPostRsp("/space/v1/api/file/copy?userId=1", nil, &req, &rsp, assert)
		assert.Equal(int(proto.CodeReqParamErr), int(rsp.Code))
	}

	{
		req.Ids = append(req.Ids, "xxxxxxxxxxxxxxxxxx")
		TPostRsp("/space/v1/api/file/copy?userId=1", nil, &req, &rsp, assert)
		assert.Equal(int(proto.CodeFailedToOperateDB), int(rsp.Code))
	}

}

func testFilesList(t *testing.T) {
	assert := assert.New(t)
	var rsp proto.Rsp
	var body proto.GetListRspData
	rsp.Body = &body
	TGetRsp("/space/v1/api/file/list?userId=1", &rsp, assert)
	assert.Equal(int(200), int(rsp.Code))
	assert.True(len(body.List) > 0)
}

func testFilesRename(t *testing.T) {
	assert := assert.New(t)
	var rsp proto.Rsp
	var req proto.ModifyFileReq
	var body proto.DbAffect
	//var rspHead proto.Rsp

	if fi, err := dbutils.GetInfoByPath(1, "/", "说明.pdf", proto.TrashStatusNormal); err == nil {
		req.Id = fi.Id
		req.NewFileName = strconv.FormatInt(time.Now().Unix(), 10)
	}
	rsp.Body = &body
	TPostRsp("/space/v1/api/file/rename?userId=1", nil, &req, &rsp, assert)
	assert.Equal(int(200), int(rsp.Code))
	assert.True(body.AffectRows > 0)
}

func testFolderRename(t *testing.T) {
	assert := assert.New(t)
	var rsp proto.Rsp
	var req proto.ModifyFileReq
	var body proto.DbAffect
	//var rspHead proto.Rsp

	if fi, err := dbutils.GetInfoByPath(1, "/", "文档", proto.TrashStatusNormal); err == nil {
		req.Id = fi.Id
		req.NewFileName = uuid.New().String()
	} else {
		assert.Equal(nil, err, "err=%v", err)
		return
	}
	rsp.Body = &body
	TPostRsp("/space/v1/api/file/rename?userId=1", nil, &req, &rsp, assert)
	assert.Equal(int(200), int(rsp.Code))
	assert.True(body.AffectRows > 0)
}

func testFilesMove(t *testing.T) {
	assert := assert.New(t)
	var rsp proto.Rsp
	var req proto.MoveFileReq
	var body proto.DbAffect
	//var rspHead proto.Rsp
	api.InitReadme(1)

	if fi, err := dbutils.GetInfoByPath(1, "/", "说明.pdf", proto.TrashStatusNormal); err == nil {
		req.Id = append(req.Id, fi.Id)
	}
	if di, err := dbutils.GetInfoByPath(1, "/", "文档", proto.TrashStatusNormal); err == nil {
		req.DestPath = di.Id
	}

	rsp.Body = &body
	TPostRsp("/space/v1/api/file/move?userId=1", nil, &req, &rsp, assert)
	assert.Equal(int(200), int(rsp.Code))
	assert.True(body.AffectRows > 0)
}

//func testFilesCopy(t *testing.T) {
//	assert := assert.New(t)
//	var rsp proto.Rsp
//	var req proto.CopyFileReq
//	var body proto.DbAffect
//	//var rspHead proto.Rsp
//	initReadme(1)
//
//	if fi, err := dbutils.GetInfoByPath(1, "/", "readme.txt", proto.TrashStatusNormal); err == nil {
//		req.Id = append(req.Id,fi.Id)
//	}
//	if di, err := dbutils.GetInfoByPath(1, "/", "Documents", proto.TrashStatusNormal); err == nil {
//		req.DestPath = di.Id
//	}
//
//
//	rsp.Body = &body
//	TPostRsp("/space/v1/api/file/copy?userId=1",nil,&req, &rsp, assert)
//	assert.Equal(int(200), int(rsp.Code))
//	assert.True(body.AffectRows > 0)
//}

func TestAsyncDelete(t *testing.T) {
	assert := assert.New(t)
	var rsp proto.Rsp
	var body async.AsyncTask
	rsp.Body = &body
	var req proto.DeleteFileReq
	req.DeleteIds = []string{"2003d06b-f7f5-4c7d-bf41-a07f19cfd249"}
	TPostRsp("/space/v1/api/file/delete?userId=1", nil, &req, &rsp, assert)
	fmt.Println(body)
	taskId := body.TaskId
	for {
		TGetRsp(fmt.Sprintf("/space/v1/api/async/task?userId=1&taskId=%s", taskId), &rsp, assert)
		if body.TaskStatus == async.AsyncTaskStatusSuccess || rsp.Code == proto.CodeGetAsyncTaskInfoFailed {
			break
		}
		time.Sleep(100 * time.Millisecond)

	}
}

func TestAsyncRestore(t *testing.T) {
	assert := assert.New(t)
	var rsp proto.Rsp
	var body async.AsyncTask
	rsp.Body = &body
	var req proto.RestoreRecycledReq
	req.RecycledUuids = []string{"2003d06b-f7f5-4c7d-bf41-a07f19cfd249"}
	TPostRsp("/space/v1/api/recycled/restore?userId=1", nil, &req, &rsp, assert)
	taskId := body.TaskId
	for {
		TGetRsp(fmt.Sprintf("/space/v1/api/async/task?userId=1&taskId=%s", taskId), &rsp, assert)
		if body.TaskStatus == async.AsyncTaskStatusSuccess || rsp.Code == proto.CodeGetAsyncTaskInfoFailed {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}
