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

package dbutils

import "aofs/internal/proto"

type betagIndex struct {
}

func (*betagIndex) Get(betag string) (int, error) {
	var bi proto.BETagInfo
	result := db.Model(proto.BETagInfo{}).Where("betag=?", betag).First(&bi)
	if result.Error != nil {
		return 0, result.Error
	} else {
		return int(bi.VolId), nil
	}
}
func (*betagIndex) Delete(betag string) (int, error) {
	trans, err := newTrans()
	if err != nil {
		return 0, err
	}

	defer trans.Commit()

	bi, err := trans.getBETagInfo(betag)
	if err != nil {
		return 0, err
	}

	err = trans.deleteBETagInfo(betag)
	return int(bi.VolId), err
}
func (*betagIndex) Add(betag string, volId int) error {
	trans, err := newTrans()
	if err != nil {
		return err
	}

	defer trans.Commit()
	return trans.addBETagInfo(betag, volId)
}

func NewBETagIndexer() *betagIndex {
	return &betagIndex{}
}
