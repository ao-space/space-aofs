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

package routers

import (
	"aofs/routers/api"
	"aofs/routers/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	gs "github.com/swaggo/gin-swagger"
)

func InitRoute() *gin.Engine {
	route := gin.New()

	//内部接口，不需要放到路由表中
	route.Use(middleware.LoggerHandler(), middleware.InitSession())

	inner := route.Group("/space/v1/api/inner")
	{
		inner.GET("/file/info", api.GetFileInfoForInner)
		inner.POST("/file/infos", api.GetFileInfosForTrends)
	}
	// 文件接口
	file := route.Group("/space/v1/api/file")
	{
		file.GET("/info", api.GetFileInfo)
		file.GET("/list", api.ListFiles)
		file.POST("/rename", api.ModifyFile)
		file.POST("/copy", api.CopyFiles)
		file.POST("/move", api.MoveFile)
		file.POST("/delete", api.TrashFiles)
		file.GET("/download", api.DownloadFile)
		file.GET("/search", api.SearchFiles)
		file.GET("/thumb", api.GetThumb)
		file.GET("/compressed", api.GetCompressed)
		file.POST("/vod/symlink", api.CreateVodSymlink)
	}

	folder := route.Group("/space/v1/api/folder")
	{
		folder.POST("/create", api.CreateFolders)
		folder.GET("/info", api.FolderInfo)
	}
	// user接口
	user := route.Group("/space/v1/api/user")
	{
		user.POST("/init", api.UserInit)
		user.POST("/delete", api.UserDelete)
		user.GET("/storage", api.UserUsedSpace)
	}
	//同步接口
	sync := route.Group("/space/v1/api/sync")
	{
		sync.GET("/synced", api.GetSyncedFiles)
	}

	// 回收站接口
	recycled := route.Group("/space/v1/api/recycled")
	{
		recycled.GET("/clear", api.ClearRecycled)
		recycled.POST("/clear", api.ClearRecycled)
		recycled.POST("/restore", api.RestoreRecycled)
		recycled.GET("/list", api.ListRecycled)

	}

	// multipart 分片上传
	multipart := route.Group("/space/v1/api/multipart")
	{
		multipart.POST("/create", api.CreateMultipartTask)
		multipart.POST("/delete", api.DeleteMultipartTask)
		multipart.GET("/list", api.ListMultipartTask)
		multipart.POST("/upload", api.UploadPart)
		multipart.POST("/complete", api.CompleteMultipartTask)
	}

	route.GET("/space/v1/api/status", api.Status)

	async := route.Group("/space/v1/api/async")
	{
		async.GET("/task", api.GetAsyncTaskInfo)
	}

	if gin.Mode() == gin.DebugMode {
		route.GET("swagger/*any", gs.WrapHandler(swaggerFiles.Handler))
	}

	return route
}
