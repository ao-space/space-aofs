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

package proto

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ReqHead2 struct {
	RequestId string `json:"requestId" form:"requestId"`
}

type Rsp struct {
	c         *gin.Context
	Code      CodeType    `json:"code" form:"code"`           //msg code
	Message   string      `json:"message" form:"message"`     //msg info
	RequestId string      `json:"requestId" form:"requestId"` //trans id
	Body      interface{} `json:"results" form:"results"`     //response body
} //@name Rsp

func (rsp *Rsp) GetCode() CodeType {
	return rsp.Code
}
func (rsp *Rsp) GetMsg() string {
	return rsp.Message
}
func (rsp *Rsp) GetBody() interface{} {
	return rsp.Body
}

func (rsp *Rsp) SetContext(c *gin.Context) {
	rsp.c = c
}

//发送正常响应OK的回应消息
//	body:回应的消息包体（消息头在函数里面自动完成)
func (rsp *Rsp) SendOk(body interface{}) {
	rsp.Code = CodeOk
	rsp.Message = GetMessageByCode(CodeOk)
	rsp.Body = body

	//此处c必须为有效对象，就不再做判断，利用系统panic
	rsp.c.JSON(http.StatusOK, rsp)
}

//发送错误响应的回应消息
//	code: 回应的错误码
//	err: 回应的错误信息扩展描述
func (rsp *Rsp) SendErr(code CodeType, err error) {
	rsp.Code = code
	rsp.Message = GetMessageByCode(CodeType(code))
	if err != nil {
		rsp.Message += "[ext: " + err.Error() + " ]"
	}
	rsp.Body = nil

	//此处c必须为有效对象，就不再做判断，利用系统panic
	rsp.c.JSON(http.StatusOK, rsp)
}
