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

package config

import (
	"os"
	"strconv"
)

//环境变量不区分大小写
func ReadString(key string, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	} else {
		return def
	}
}

func ReadBool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		return v == "1"
	} else {
		return def
	}
}

func ReadInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok {
		if intV, err := strconv.Atoi(v); err == nil {
			return intV
		}
	}
	return def
}

func ReadInt64(key string, def int64) int64 {
	if v, ok := os.LookupEnv(key); ok {
		if intV, err := strconv.ParseInt(v, 10, 64); err == nil {
			return intV
		}
	}
	return def
}
