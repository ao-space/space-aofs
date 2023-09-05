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

package proto

import (
	"fmt"
)

/*
此文件定义回应码和描述信息
*/

//错误码在此部分定义
type CodeType uint32

func (ct CodeType) String() string {
	return fmt.Sprintf("CODE_%d", ct)
}

const (
	CodeOk                     CodeType = 200
	CodeCreateAsyncTaskSuccess CodeType = 201

	CodeParamErr                    CodeType = 1001 //参数错误
	CodeReqParamErr                 CodeType = 1002 //请求参数错误(推荐)
	CodeFileNotExist                CodeType = 1003 //文件不存在
	CodeFailedToOpenFile            CodeType = 1004 //打开文件失败
	CodeFolderNotExist              CodeType = 1005 //文件夹不存在
	CodeFailedToConnect             CodeType = 1006 //建立连接失败
	CodeFailedToConnectDB           CodeType = 1007 //连接DB失败
	CodeFailedToOptMinio            CodeType = 1008 //失败去写文件
	CodeFailedToCreateBucket        CodeType = 1009 //失败去创建桶
	CodeFailedToSaveFile            CodeType = 1010 //失败去保存数据
	CodeFailedToOperateDB           CodeType = 1011 //失败去操作数据库
	CodeFailedToCreateFolder        CodeType = 1012 //失败去创建文件夹
	CodeFileExist                   CodeType = 1013 //文件已存在
	CodeFolderExist                 CodeType = 1014 //文件夹已存在
	CodeFolderDepthTooLong          CodeType = 1015 //文件夹层数超过5层
	CodeUserIdError                 CodeType = 1016 //用户Id的错误
	CodeFailedToInitUser            CodeType = 1017 //失败去初始化用户
	CodeFailedToDeleteUser          CodeType = 1018 //删除用户失败
	CodeFailedToGetUsedStorage      CodeType = 1021 //获取用户空间使用量失败
	CodeCopyIdError                 CodeType = 1022 //文件夹不能拷贝到自己

	CodeFailedToCreateMultipartTask CodeType = 1027 // 失败去创建任务
	CodeMultipartTaskNotFound       CodeType = 1028 //未找到任务
	CodeMultipartTaskOverlap        CodeType = 1029 //任务重叠
	CodeMultipartTaskRangeErr       CodeType = 1030 //上传范围错误
	CodeMultipartTaskHashErr        CodeType = 1031 //文件hash校验错误
	CodeMultipartTaskCompleteErr    CodeType = 1032 //合并错误
	CodeNotEnoughSpace              CodeType = 1036 //空间不够，不上传
	CodeMultipartRangeUploaded      CodeType = 1037 //分片范围已上传
	CodeMultipartUploadingConflit   CodeType = 1038 //分片范围上传冲突

	CodeWriteRedisFailed       CodeType = 1046 //写入redis 失败

	CodeZipFileFailed          CodeType = 1050 //压缩文件失败

	CodeFailedToCreateSymlink  CodeType = 1061 //失败去创建符号链接
	CodeGetAsyncTaskInfoFailed CodeType = 1062 // 获取异步任务状态失败
)

//错误码对应描述在此部分定义
var codeMessageMap map[CodeType]string

func init() {
	codeMessageMap = map[CodeType]string{}

	//开始定义
	codeMessageMap[CodeOk] = "OK"
	codeMessageMap[CodeFileNotExist] = "File is not exist"
	codeMessageMap[CodeFolderNotExist] = "Folder is not exist"
	codeMessageMap[CodeFailedToOperateDB] = "failed to operate db"
	codeMessageMap[CodeFailedToSaveFile] = "failed to save file"
	codeMessageMap[CodeFileExist] = "File already exists"
	codeMessageMap[CodeFolderExist] = "Folder already exists "
	codeMessageMap[CodeFolderDepthTooLong] = "Folder Depth Is Too Long! Failed to Create Folder"
	codeMessageMap[CodeUserIdError] = "User Id Error！Should Be Greater than or equal 1"
	codeMessageMap[CodeCopyIdError] = "File Operation: DestPath could not be itself"
	codeMessageMap[CodeNotEnoughSpace] = "Normal Upload: not enough space"
}

// GetMessageByCode 根据错误码获取描述
func GetMessageByCode(code CodeType) string {
	if message, ok := codeMessageMap[code]; ok {
		return message
	} else {
		return ""
	}
}

type ErrMess struct {
	Code    CodeType `json:"code"`
	Message string   `json:"message"`
}

type BpErr struct {
	Code CodeType
	Err  error
}

const (
	FileNotFound     string = "FileNotFound"
	S3OperationError string = "S3OperationError"
	ArgsErrors       string = "ArgsErrors"
)
