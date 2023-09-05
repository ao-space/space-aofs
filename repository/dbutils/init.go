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

import (
	"aofs/internal/env"
	"aofs/internal/log4bp"
	"aofs/internal/proto"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB
var db2 *gorm.DB
var geodb *sql.DB
var logdb = log4bp.New("", gin.Mode())

func SetMockDb(mockdb *gorm.DB) {
	db = mockdb
}

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	Config   *gorm.Config
}

func GormLogger() *logger.Interface {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer（日志输出的目标，前缀和日志包含的内容）
		logger.Config{
			SlowThreshold:             time.Millisecond * 200, // 慢 SQL 阈值
			LogLevel:                  logger.Error,           // 日志级别
			IgnoreRecordNotFoundError: true,                   // 忽略ErrRecordNotFound（记录未找到）错误
			Colorful:                  false,                  // 禁用彩色打印
		},
	)
	return &newLogger
}

func FileSchemaConn(conf *DBConfig) *gorm.DB {
	connArgs := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable", conf.Host, conf.Port, conf.User, conf.Password, conf.DBName)
	db, err := gorm.Open(postgres.Open(connArgs), conf.Config)
	if err != nil {
		return nil
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxIdleTime(time.Hour)
	return db
}

func AccountSchemaConn(conf *DBConfig) *gorm.DB {
	connArgs := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable", conf.Host, conf.Port, conf.User, conf.Password, conf.DBName)
	db2, err := gorm.Open(postgres.Open(connArgs), conf.Config)
	if err != nil {
		return nil
	}
	return db2
}


func CreateTable(table interface{}) {
	err := db.AutoMigrate(table)
	if err != nil {
		logdb.LogF().Err(err).Msg("failed to create table")
		panic(any(err))
	}
}

func Init() {
	createTables()
}

func InitUser(userId proto.UserIdType) {

	folderRoot := proto.FileInfo{
		FileInfoPub: proto.FileInfoPub{
			Id:            uuid.NewString(),
			IsDir:         true,
			Name:          "/",
			Path:          "",
			BETag:         "",
			CreateTime:    time.Now().UnixNano() / 1e6,
			ModifyTime:    time.Now().UnixNano() / 1e6,
			OperationTime: time.Now().UnixNano() / 1e6,
			Size:          0,
			FileCount:     1,
			Category:      "",
			Mime:          "",
		},

		UserId: userId,

		Tags: "",

		Executable:    false,
		Version:       1,
		BucketName:    "eulixspace-files",
		TransactionId: 0,
	}
	folderDocuments := proto.FileInfo{
		FileInfoPub: proto.FileInfoPub{
			Id:            uuid.NewString(),
			IsDir:         true,
			Name:          "文档",
			Path:          "/",
			BETag:         "",
			CreateTime:    time.Now().UnixNano() / 1e6,
			ModifyTime:    time.Now().UnixNano() / 1e6,
			OperationTime: time.Now().UnixNano() / 1e6,
			Size:          0,
			Category:      "",
			Mime:          "",
		},

		UserId: userId,
		Tags:   "",

		Executable: false,

		Version:       1,
		BucketName:    "eulixspace-files",
		TransactionId: 0,
	}
	folderVideos := proto.FileInfo{
		FileInfoPub: proto.FileInfoPub{
			Id:            uuid.NewString(),
			IsDir:         true,
			Name:          "视频",
			Path:          "/",
			BETag:         "",
			CreateTime:    time.Now().UnixNano() / 1e6,
			ModifyTime:    time.Now().UnixNano() / 1e6,
			OperationTime: time.Now().UnixNano() / 1e6,
			Size:          0,
			Category:      "",
			Mime:          "",
		},

		UserId:        userId,
		Tags:          "",
		Executable:    false,
		Version:       1,
		BucketName:    "eulixspace-files",
		TransactionId: 0,
	}
	folderPhotos := proto.FileInfo{
		FileInfoPub: proto.FileInfoPub{
			Id:            uuid.NewString(),
			IsDir:         true,
			Name:          "图片",
			Path:          "/",
			BETag:         "",
			CreateTime:    time.Now().UnixNano() / 1e6,
			ModifyTime:    time.Now().UnixNano() / 1e6,
			OperationTime: time.Now().UnixNano() / 1e6,
			Size:          0,
			Category:      "",
			Mime:          "",
		},

		UserId: userId,

		Tags: "",

		Executable: false,

		Version:       1,
		BucketName:    "eulixspace-files",
		TransactionId: 0,
	}

	VerifyInit(userId, folderRoot.Name, folderRoot.Path, folderRoot)
	VerifyInit(userId, folderDocuments.Name, folderDocuments.Path, folderDocuments)
	VerifyInit(userId, folderVideos.Name, folderVideos.Path, folderVideos)
	VerifyInit(userId, folderPhotos.Name, folderPhotos.Path, folderPhotos)

}

func VerifyInit(uid proto.UserIdType, name string, path string, info proto.FileInfo) {

	rows := db.Where("name = ? AND path = ? AND user_id = ?", name, path, uid).First(&proto.FileInfo{})
	if rows.RowsAffected != 1 {
		db.Create(&info)
	}
}


func GetFileDB() *gorm.DB {
	return db
}

func GetAccountDB() *gorm.DB {
	return db2
}
func createTables() {
	fileConnConfig := &DBConfig{
		Host:     env.SQL_HOST,
		Port:     env.SQL_PORT,
		User:     env.SQL_USER,
		Password: env.SQL_PASSWORD,
		DBName:   env.SQL_DATABASE,
		Config:   &gorm.Config{Logger: *GormLogger(), PrepareStmt: true},
	}

	accountConnConfig := &DBConfig{
		Host:     env.SQL_HOST,
		Port:     env.SQL_PORT,
		User:     env.SQL_USER,
		Password: env.SQL_PASSWORD,
		DBName:   "account",
		Config:   &gorm.Config{Logger: *GormLogger(), PrepareStmt: true},
	}

	db = FileSchemaConn(fileConnConfig)
	db2 = AccountSchemaConn(accountConnConfig)
	geodb, _ = db.DB()

	// 建表
	logdb.LogI().Msg(fmt.Sprintf("Connected Database:%v:%d/%v", env.SQL_HOST, env.SQL_PORT, env.SQL_DATABASE))
	CreateTable(proto.Setting{})
	CreateTable(proto.BETagInfo{})
	CreateTable(proto.FileInfo{})
	CreateTable(proto.SyncInfo{})

}
