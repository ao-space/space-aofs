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

package bpctx

import (
	"aofs/internal/log4bp"
	"aofs/internal/proto"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

type Context struct {
	//zerolog.Logger
	s      time.Time //开始时间
	c      *gin.Context
	tid    string
	userId proto.UserIdType `validate:"gte=1"`
	rsp    proto.Rsp
	m      map[string]interface{}
}

func (ctx *Context) GetContext() *gin.Context {
	return ctx.c
}

func (ctx *Context) GetRawLogger() *zerolog.Logger {
	return &log4bp.Logger
}

func (ctx *Context) Append(key string, value interface{}) {
	ctx.m[key] = value
}

func NewCtx(c *gin.Context) *Context {
	ctx := &Context{c: c, s: time.Now(), m: make(map[string]interface{})}
	if c != nil {
		ctx.getUserId()
		ctx.getTid()
	}

	return ctx
}

func NewCtxForTest(userId proto.UserIdType) *Context {
	ctx := &Context{c: &gin.Context{}}
	ctx.userId = userId

	ctx.getTid()

	return ctx
}

func (ctx *Context) LogI(title string, req interface{}) {
	var l *zerolog.Event
	if ctx.rsp.Code == proto.CodeOk {
		l = log4bp.Logger.Info()
	} else {
		l = log4bp.Logger.Warn()
	}
	l = l.Timestamp().Caller(1).Str("requestId", ctx.tid).Int("userId", int(ctx.userId)).Str("title", title).Interface("req", req).Interface("rsp", ctx.rsp.GetBody()).Int64("cost", int64(time.Now().Sub(ctx.s)/1e6)).Interface("code", ctx.rsp.Code)
	for k, v := range ctx.m {
		l = l.Interface(k, v)
	}
	l.Msg(ctx.rsp.Message)
}

func (ctx *Context) LogD() *zerolog.Event {
	return log4bp.Logger.Debug().Timestamp().Caller(1).Str("tid", ctx.tid).Int("userId", int(ctx.userId))
}
func (ctx *Context) LogW() *zerolog.Event {
	return log4bp.Logger.Warn().Timestamp().Caller(1).Str("tid", ctx.tid).Int("userId", int(ctx.userId))
}
func (ctx *Context) LogE() *zerolog.Event {
	return log4bp.Logger.Error().Timestamp().Caller(1).Str("tid", ctx.tid).Int("userId", int(ctx.userId))
}

//获取 userId
func (ctx *Context) GetUserId() proto.UserIdType {
	return ctx.userId
}

func (ctx *Context) getUserId() {
	userId := ctx.c.Query("userId")
	u64, err := strconv.ParseUint(userId, 10, 8)
	if err != nil || u64 == 0 {
		//校验参数错误
		ctx.SendErr(proto.CodeParamErr, err)
		panic(any(u64))
	} else {
		ctx.userId = proto.UserIdType(u64)
	}
}

func (ctx *Context) getTid() string {
	if ctx.tid == "" {
		ctx.tid = ctx.c.Query("requestId")
		if ctx.tid == "" {
			ctx.tid = uuid.New().String()[:13]
		}
	}
	return ctx.tid
}

//发送错误响应的回应消息
//	code: 回应的错误码
//	err: 回应的错误信息扩展描述
func (ctx *Context) SendErr(code proto.CodeType, err error) {
	if ctx.c.Request.Body != nil {
		io.Copy(ioutil.Discard, ctx.c.Request.Body)
	}

	ctx.rsp.Code = code
	ctx.rsp.Message = proto.GetMessageByCode(code)
	if err != nil {
		ctx.rsp.Message += "[ext: " + err.Error() + " ]"
	}
	ctx.rsp.Body = nil

	//此处c必须为有效对象，就不再做判断，利用系统panic
	if ctx.c != nil {
		ctx.c.JSON(http.StatusOK, ctx.rsp)
	}

}

//发送正常响应OK的回应消息
//	body:回应的消息包体（消息头在函数里面自动完成)
func (ctx *Context) SendOk(body interface{}) {
	if ctx.c.Request.Body != nil {
		io.Copy(ioutil.Discard, ctx.c.Request.Body)
	}

	ctx.rsp.Code = proto.CodeOk
	ctx.rsp.Message = proto.GetMessageByCode(proto.CodeOk)
	ctx.rsp.Body = body

	//此处c必须为有效对象，就不再做判断，利用系统panic
	ctx.c.JSON(http.StatusOK, ctx.rsp)
}

func (ctx *Context) SendRsp(body interface{}, bperr *proto.BpErr) {
	if ctx.c.Request.Body != nil {
		io.Copy(ioutil.Discard, ctx.c.Request.Body)
	}

	ctx.rsp.Code = bperr.Code
	ctx.rsp.Message = proto.GetMessageByCode(bperr.Code)
	ctx.rsp.Body = body

	//此处c必须为有效对象，就不再做判断，利用系统panic
	ctx.c.JSON(http.StatusOK, ctx.rsp)
}

func (ctx *Context) Send(httpCode proto.CodeType, body interface{}) {
	if ctx.c.Request.Body != nil {
		io.Copy(ioutil.Discard, ctx.c.Request.Body)
	}

	ctx.rsp.Code = httpCode
	ctx.rsp.Message = proto.GetMessageByCode(httpCode)
	ctx.rsp.Body = body

	//此处c必须为有效对象，就不再做判断，利用系统panic
	ctx.c.JSON(int(httpCode), ctx.rsp)
}
