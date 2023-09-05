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
	"fmt"
	"testing"
	"time"

	"gorm.io/gorm"
)

func init() {
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

	randBetag = fmt.Sprint(time.Now().Unix())
}

var randBetag string

var bi betagIndex

func testBETagIndexGet(t *testing.T) {
	type args struct {
		betag string
	}
	tests := []struct {
		name    string
		b       *betagIndex
		args    args
		want    int
		wantErr bool
	}{
		{name: "get-add1", b: &bi, args: args{betag: randBetag}, want: 1, wantErr: false},
		{name: "get-add1", b: &bi, args: args{betag: "not-exist"}, want: 0, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.b.Get(tt.args.betag)
			if (err != nil) != tt.wantErr {
				t.Errorf("betagIndex.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("betagIndex.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func testBETagIndexAdd(t *testing.T) {
	type args struct {
		betag string
		volId int
	}
	var bi betagIndex
	tests := []struct {
		name    string
		b       *betagIndex
		args    args
		wantErr bool
	}{
		{name: "add1", b: &bi, args: args{betag: randBetag, volId: 1}, wantErr: false},
		{name: "add1-again", b: &bi, args: args{betag: randBetag, volId: 1}, wantErr: true},
		{name: "add2", b: &bi, args: args{betag: randBetag + "-2", volId: 1}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.b.Add(tt.args.betag, tt.args.volId); (err != nil) != tt.wantErr {
				t.Errorf("betagIndex.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func testBETagIndexDelete(t *testing.T) {
	type args struct {
		betag string
	}
	var bi betagIndex
	tests := []struct {
		name    string
		b       *betagIndex
		args    args
		want    int
		wantErr bool
	}{
		{name: "delete add1", b: &bi, args: args{betag: randBetag}, want: 1, wantErr: false},
		{name: "delete add1-again", b: &bi, args: args{betag: randBetag}, want: 0, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.b.Delete(tt.args.betag)
			if (err != nil) != tt.wantErr {
				t.Errorf("betagIndex.Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("betagIndex.Delete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBETagIndex(t *testing.T) {
	testBETagIndexAdd(t)
	testBETagIndexGet(t)
	testBETagIndexDelete(t)
}
