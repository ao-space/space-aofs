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

	"github.com/gin-gonic/gin"
)

// GetSyncedFiles  Obtain incremental synchronization data
// @Summary  Obtain incremental synchronization data
// @Description Obtain incremental synchronization data
// @Tags Sync
// @Accept application/json
// @Produce application/json
// @Param userId query string true "user id"
// @Param timestamp query int true "timestamp"
// @Param path query string false "path"
// @Param deviceId query string true "device id"
// @Success 200 {object} proto.Rsp{results=proto.GetListRspData} ""
// @Router /space/v1/api/sync/synced [GET]
func GetSyncedFiles(c *gin.Context) {

	ctx := bpctx.NewCtx(c)

	var syncedFilesReq proto.SyncedFilesReq

	var rspData proto.GetListRspData


	userId := ctx.GetUserId()

	if err := c.ShouldBind(&syncedFilesReq); err != nil {
		ctx.SendErr(proto.CodeReqParamErr, err)
		return
	}

	changedList, syncErr := dbutils.GetSyncedFiles(syncedFilesReq.DeviceId, syncedFilesReq.Path, syncedFilesReq.Timestamp, userId)
	if syncErr != nil {
		ctx.SendErr(proto.CodeFileNotExist, syncErr)
		return
	} else {
		rspData.List = changedList.ToPubLst()
		ctx.SendOk(&rspData)
		return
	}

}
