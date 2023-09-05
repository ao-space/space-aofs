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

package bpredis

import (
	"time"
)

const UsedSpace = "UsedSpace-"

func (br *bpRedis) Set(key string, value interface{}, expiration time.Duration) error {
	err := br.Client.Set(key, value, expiration).Err()
	if err != nil {
		return err
	}
	return nil
}

func (br *bpRedis) Get(key string) (int, error) {
	uv, err := br.Client.Get(key).Int()
	if err != nil {
		return 0, err
	}
	return uv, nil
}

func (br *bpRedis) GetInt64(key string) (int64, error) {
	uv, err := br.Client.Get(key).Int64()
	if err != nil {
		return 0, err
	}
	return uv, nil
}

func (br *bpRedis) GetValue(key string) ([]byte, error) {
	value := br.Client.Get(key)
	if value.Err() != nil {
		return nil, value.Err()
	}
	rspByte, err := value.Bytes()
	if err != nil {
		return nil, err
	}

	return rspByte, nil
}

func (br *bpRedis) Incr(key string) error {
	err := br.Client.Incr(key).Err()
	if err != nil {
		return err
	}
	return nil
}
