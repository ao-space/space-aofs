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

package file

import (
	"encoding/base64"
	"encoding/json"
	"aofs/internal/proto"
	"aofs/repository/bpredis"
	"aofs/repository/dbutils"
	"github.com/google/uuid"
)

var redis = bpredis.GetRedis()

func PushChanges(optType string, userId proto.UserIdType, uuids []string) error {
	var changeUuids []string
	files, err := dbutils.GetFilesInUuids(userId, uuids)
	if err != nil {
		return err
	}

	subfiles, err := dbutils.GetSubFilesInUuids(userId, uuids, []uint32{proto.TrashStatusNormal, proto.TrashStatusLogicDeleted, proto.TrashStatusSubFilesLogicDeleted})
	if err != nil {
		return err
	}

	changeUuids = append(changeUuids, files...)

	if len(subfiles) > 0 {
		changeUuids = append(changeUuids, subfiles...)
	}

	pushData := &proto.FileChangePushMsg{
		OperatorType: optType,
		Uuids:        changeUuids,
	}
	dataBytes, _ := json.Marshal(pushData)
	fileChangeMsg := map[string]interface{}{
		"userId":    int(userId),
		"optType":   "HISTORY_RECORD",
		"requestId": uuid.New().String(),
		"data":      base64.StdEncoding.EncodeToString(dataBytes),
	}
	redis.PushNotificationMsg(fileChangeMsg)

	return nil
}
