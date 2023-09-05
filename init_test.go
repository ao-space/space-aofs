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

package main

import (
	"aofs/internal/log4bp"
	"aofs/repository/dbutils"
	"database/sql"
	"log"

	"github.com/rs/zerolog"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var mock sqlmock.Sqlmock

//初始化
func init() {
	//calcFile(`C:\Users\michael\Downloads\IMG_5585.MOV`)
	//testMultipartRandFun(6*1024*1024, nil)
	//mock 数据库
	//创建sqlmock
	var err error
	var dbmock *sql.DB
	dbmock, mock, err = sqlmock.New()
	if nil != err {
		log.Fatalf("Init sqlmock failed, err %v", err)
	}
	//结合gorm、sqlmock
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: dbmock}), &gorm.Config{})

	if nil != err {
		log.Fatalf("Init DB with sqlmock failed, err %v", err)
	}

	//暂时不启用mock
	if false {
		dbutils.SetMockDb(db)
	}

	//初始化系统数据库
	//Init()
	log4bp.Logger.Level(zerolog.DebugLevel)
}
