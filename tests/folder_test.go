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
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func testFolderAll(t *testing.T) {
	t.Run("testFolderCreate", testFolderCreate)
}

func testFolderCreate(t *testing.T) {
	assertions := assert.New(t)
	var rsp proto.Rsp
	var req proto.CreateFolderReq

	var body proto.FileInfo
	req.FolderName = strconv.FormatInt(time.Now().UnixNano(), 10)
	req.CurrentDirUuid = ""
	rsp.Body = &body

	TPostRsp("/space/v1/api/folder/create?userId=1", nil, &req, &rsp, assertions)

	assertions.Equal(int(200), int(rsp.Code))
	assertions.True(body.Name == req.FolderName, "req:%v, rsp:%v", req.FolderName, body.Name)
}

// func testFolderInfo(t *testing.T) {
// 	assert := assert.New(t)
// 	var rsp proto.Rsp
// 	var body proto.FolderInfo
// 	var req proto.FolderInfoReq

// 	// 上传一个文件
// 	t.Run("testUpload", testUpload)

// 	if fi, err := dbutils.GetInfoByPath(1, "/", "Documents", proto.TrashStatusNormal); err == nil {
// 		req.FolderUuid = fi.Id
// 	}

// 	TPostRsp("/space/v1/api/folder/info?userId=1", nil, &req, &rsp, assert)
// 	rsp.Body = &body

// 	assert.Equal(int(200), int(rsp.Code))
// 	assert.True(body.FolderSize > 0)
// }

// func TestMoveFile(t *testing.T) {

// 	reqstr := `
// 	{
//     	"id": "91c8c5c3-ad46-4721-9687-09df8f2e88c2",
//     	"dest-path": "/picture"
// 	}
// 	`

// 	if rsp, err := http.Post(fmt.Sprintf("http://%s/space/v1/api/files/move", addr), "application/json", strings.NewReader(reqstr)); err != nil {
// 		t.Error(err)
// 	} else {
// 		defer rsp.Body.Close()
// 		io.Copy(os.Stdout, rsp.Body)
// 	}
// }
