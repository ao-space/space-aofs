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

	"github.com/gin-gonic/gin"
)

// Status Querying service status
// @Summary Querying service status
// @Description Querying service status
// @Tags ServerStatus
// @Accept application/json
// @Produce application/json
// @Param userId query string true "user id"
// @Success 200 {object} proto.Rsp{results=proto.StatusRsp} "success"
// @Router /space/v1/api/status [GET]
func Status(c *gin.Context) {

	ctx := bpctx.NewCtx(c)

	var statusRsp proto.StatusRsp

	defer ctx.LogI("get status", nil)

	statusRsp.Status = "ok"
	ctx.SendOk(&statusRsp)

}
