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

package api

import (
	"aofs/internal/bpctx"
	"aofs/internal/proto"
	"aofs/repository/dbutils"
	"aofs/services/recycled"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserInit Initialize user
// @Summary Initialize user
// @Description Initialize user
// @Tags User
// @Accept application/json
// @Produce application/json
// @Param userId query string true "user id"
// @Param userId body proto.User true "user id"
// @Success 200 {object} proto.Rsp ""
// @Router /space/v1/api/user/init [POST]
func UserInit(c *gin.Context) {

	ctx := bpctx.NewCtx(c)

	defer ctx.LogI("init user", ctx.GetUserId())
	if err := initUser(ctx.GetUserId()); err != nil {
		ctx.SendErr(proto.CodeFailedToInitUser, err)
	} else {
		ctx.SendOk(nil)
	}

}

// UserUsedSpace Query the user space capacity
// @Summary Query the user space capacity
// @Description Query the user space capacity
// @Tags User
// @Accept application/json
// @Produce application/json
// @Param userId query string true "user id"
// @Param targetUserId query string true "user id"
// @Success 200 {object} proto.Rsp{results=proto.Storage} ""
// @Router /space/v1/api/user/storage [GET]
func UserUsedSpace(c *gin.Context) {

	ctx := bpctx.NewCtx(c)

	var storage proto.Storage
	userId := ctx.GetUserId()

	targetUserId := c.Query("targetUserId")
	u64, err := strconv.ParseUint(targetUserId, 10, 64)

	defer ctx.LogI("Get used space", userId)
	if err != nil || u64 == 0 {
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}
	if userId >= 1 && u64 >= 1 && userId <= proto.UserIdType(u64) {
		if userId == 1 || userId == proto.UserIdType(u64) {
			if res, err := dbutils.GetUsedSpaceByUser(proto.UserIdType(u64)); err != nil {
				storage.UserStorage = 0
				ctx.SendOk(&storage)
				//ctx.SendErr(proto.CodeFailedToGetUsedStorage, err)
				return
			} else {
				storage.UserStorage = res
				ctx.SendOk(&storage)
			}
		}
	} else {
		ctx.SendErr(proto.CodeUserIdError, nil)
	}

}

// UserDelete Delete user
// @Summary Delete user
// @Description Delete user
// @Tags User
// @Accept application/json
// @Produce application/json
// @Param userId query string true "user id"
// @Param userId body proto.User true "user id"
// @Success 200 {object} proto.Rsp ""
// @Router /space/v1/api/user/delete [POST]
func UserDelete(c *gin.Context) {
	ctx := bpctx.NewCtx(c)

	var user proto.User
	//获取请求参数

	if err := c.ShouldBind(&user); err != nil {
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}
	userId := ctx.GetUserId()

	if user.User > 1 && userId >= 1 {
		if user.User == userId || userId == 1 {
			if err := dbutils.DeleteUser(user.User); err != nil {
				ctx.SendErr(proto.CodeFailedToDeleteUser, err)
			} else {
				go recycled.DoClearRecycledTask()
				ctx.SendOk(nil)
			}
		} else {
			ctx.SendErr(proto.CodeUserIdError, nil)
			return
		}
	} else {
		ctx.SendErr(proto.CodeFailedToDeleteUser, nil)
	}

}
