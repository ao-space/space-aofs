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
	"aofs/internal/log4bp"
	"aofs/internal/proto"
	"aofs/repository/dbutils"
	"aofs/repository/storage"
	"aofs/routers/api"
	"aofs/routers/routers"
	"aofs/services/multipart"
	"aofs/services/recycled"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/rs/zerolog"

	"github.com/stretchr/testify/assert"
)

var route = routers.InitRoute()

func TGet(uri string, header http.Header) *httptest.ResponseRecorder {
	// 构造get请求
	req := httptest.NewRequest("GET", uri, nil)
	if len(header) > 0 {
		for k, v := range header {
			req.Header[k] = v
		}
	}
	// 初始化响应
	w := httptest.NewRecorder()

	// 调用相应的handler接口
	route.ServeHTTP(w, req)
	return w
}

func TPost(uri string, header http.Header, r io.Reader) *httptest.ResponseRecorder {
	// 构造get请求
	req := httptest.NewRequest("POST", uri, r)
	if header != nil {
		for k, v := range header {
			req.Header[k] = v
		}
	}
	log.Println("head", req.Header)

	// 初始化响应
	w := httptest.NewRecorder()

	// 调用相应的handler接口
	route.ServeHTTP(w, req)
	return w
}

//解析响应体到标准消息形式
func TParseRsp(response *httptest.ResponseRecorder, rsp *proto.Rsp, assert *assert.Assertions) {
	if response.Code < 200 || response.Code > 599 {
		assert.FailNowf("http code err", "response.Code:%v, expected: 200", response.Code)
	}

	defer response.Result().Body.Close()
	if data, err := ioutil.ReadAll(response.Body); err != nil {
		assert.FailNow("read body err.", "msg:%v", err)
	} else {
		if len(data) < 512 {
			fmt.Println("rsp data:", string(data))
		}
		if err := json.Unmarshal(data, &rsp); err != nil {
			assert.FailNow("parse body err.", "msg:%v", err)
		}
	}
}

//获取url的包体数据，就像下载文件一样
func TGetFile(uri string, t *testing.T) []byte {
	return TGetFileRange(uri, nil, t)
}
func TGetFileRange(uri string, part *proto.Part, t *testing.T) []byte {
	assert := assert.New(t)
	header := http.Header{}
	if part != nil {
		if part.End != -1 {
			header["Range"] = []string{fmt.Sprintf("bytes=%d-%d", part.Start, part.End)}
		} else {
			header["Range"] = []string{fmt.Sprintf("bytes=%d-", part.Start)}
		}
	}
	response := TGet(uri, header)

	if part != nil {
		if response.Code != 206 {
			assert.FailNowf("http code err", "response.Code:%v, expected: 206", response.Code)
		}
	} else {
		if response.Code != 200 {
			assert.FailNowf("http code err", "response.Code:%v, expected: 200", response.Code)
		}
	}

	defer response.Result().Body.Close()
	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		assert.FailNow("read body err:", "%v", err)
	}
	return data
}

//直接根据(GET)url到协议的标准响应json对应的响应结果
func TGetRsp(url string, rsp *proto.Rsp, assert *assert.Assertions) {
	response := TGet(url, nil)
	TParseRsp(response, rsp, assert)
}

//直接POST并得到标准协议结果, r 支持 io.Reader 或其它类型，其它类型会转成json格式上传
func TPostRsp(uri string, header http.Header, r interface{}, rsp *proto.Rsp, assert *assert.Assertions) {
	if header == nil {
		header = http.Header{}
	}
	var realr io.Reader
	if rr, ok := r.(io.Reader); ok {
		realr = rr

	} else {
		if data, err := json.Marshal(r); err != nil {
			assert.Fail("%v", err)
		} else {
			fmt.Println("req body json:", string(data))
			realr = bytes.NewReader(data)
			header["Content-Type"] = []string{"application/json"}
		}
	}

	response := TPost(uri, header, realr)
	TParseRsp(response, rsp, assert)
}

//计算md5
func TMd5sum(d []byte) string {
	ms := md5.Sum(d)
	return hex.EncodeToString(ms[:])
}

func init() {
	log4bp.Logger.Level(zerolog.DebugLevel)

	dbutils.Init()

	if err := storage.Init(dbutils.NewBETagIndexer()); err != nil {
		fmt.Println("failed to storage.Init")
		os.Exit(2)
	}

	api.Init()
	recycled.Init()  //回收站初始化
	multipart.Init() //初始化分片上传的信息

}
