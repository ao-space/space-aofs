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

package env

import "aofs/internal/config"

var (
	SQL_HOST     string // eulixspace-postgresql
	SQL_PORT     int    // 5432
	SQL_USER     string // postgres
	SQL_PASSWORD string // mysecretpassword
	SQL_DATABASE string // file



	DATA_PATH               string //数据目录,桶为此路径下的子目录
	SHARED_PATH             string //共享目录，包括磁盘信息文件
	HEAD_IMAGE_PATH         string //头像目录

	NORMAL_BUCKET string // "eulixspace-files"
	DERIVE_BUCKET string // "eulixspace-files-processed"

	//APP_BOX_PASSCODE   string //安全密码
	GIN_MODE                  string //debug or release
	RESERVED_SPACE            int64  // reserved space
	MULTIPART_TASK_LRU_SECOND int    //分片上传任务最近未使用淘汰时间，单位秒
	REDIS_URL                 string
	REDIS_PASS                string
	REDIS_DB                  int
	REDIS_STREAM_NAME         string
	BACKUP_RESERVED_SPACE     int64
	ASYNC_TASK_THRESHOLD      int
	APP_BOX_DEPLOY_METHOD     string
)

func init() {
	SQL_HOST = config.ReadString("SQL_HOST", "eulixspace-postgresql")
	SQL_PORT = config.ReadInt("SQL_PORT", 5432)
	SQL_USER = config.ReadString("SQL_USER", "postgres")
	SQL_PASSWORD = config.ReadString("SQL_PASSWORD", "mysecretpassword")
	SQL_DATABASE = config.ReadString("SQL_DATABASE", "file")
	REDIS_URL = config.ReadString("REDIS_URL", "eulixspace-redis:6379")
	REDIS_PASS = config.ReadString("REDIS_PASS", "WWQ1YZT5a4KsUI9B")
	REDIS_DB = config.ReadInt("REDIS_DB", 0)
	REDIS_STREAM_NAME = config.ReadString("REDIS_STREAM_NAME", "fileChangelogs")

	DATA_PATH = config.ReadString("DATA_PATH", "/data/")
	SHARED_PATH = config.ReadString("SHARED_PATH", "/shared/")
	HEAD_IMAGE_PATH = config.ReadString("HEAD_IMAGE_PATH", "/opt/eulixspace/image/")
	NORMAL_BUCKET = config.ReadString("MINIO_NORMAL_BUCKET", "eulixspace-files")
	DERIVE_BUCKET = config.ReadString("MINIO_DERIVE_BUCKET", "eulixspace-files-processed")

	GIN_MODE = config.ReadString("GIN_MODE", "debug")
	RESERVED_SPACE = config.ReadInt64("RESERVED_SPACE", 4831838208)
	BACKUP_RESERVED_SPACE = config.ReadInt64("BACKUP_RESERVED_SPACE", 104857600)
	MULTIPART_TASK_LRU_SECOND = config.ReadInt("MULTIPART_TASK_LIFECYCLE", 30*86400)
	ASYNC_TASK_THRESHOLD = config.ReadInt("ASYNC_TASK_THRESHOLD", 1000)
	APP_BOX_DEPLOY_METHOD = config.ReadString("APP_BOX_DEPLOY_METHOD", "box")
}
