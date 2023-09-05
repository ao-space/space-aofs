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

package api

import (
	"aofs/internal/bpctx"
	"aofs/internal/env"
	"aofs/internal/proto"
	"aofs/repository/dbutils"
	"aofs/repository/storage"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func DecodeRange(rangestr string) (*proto.Part, error) {
	if len(rangestr) == 0 {
		return &proto.Part{}, nil
	}
	rangestr = strings.TrimSpace(rangestr)
	if idx := strings.Index(rangestr, "bytes="); idx != 0 {
		return nil, fmt.Errorf("range err.%s", rangestr)
	} else {
		rangestr = rangestr[len("bytes="):]
		ranges := strings.Split(rangestr, "-")
		if len(ranges) != 2 {
			return nil, fmt.Errorf("range err.%s", rangestr)
		}
		part := &proto.Part{}
		if len(ranges[0]) > 0 {
			if i, err := strconv.Atoi(ranges[0]); err != nil || i < 0 {
				return nil, fmt.Errorf("range start err.%s", rangestr)
			} else {
				part.Start = int64(i)
			}
		}

		if len(ranges[1]) > 0 {
			if i, err := strconv.Atoi(ranges[1]); err != nil || i < 0 {
				return nil, fmt.Errorf("range end err.%s", rangestr)
			} else {
				part.End = int64(i)
			}
		} else {
			part.End = -1
		}
		if part.End > 0 && part.Start > part.End {
			return nil, fmt.Errorf("range end err.%s", rangestr)
		}

		return part, nil
	}
}

// @Summary File download
// @Description File download
// @Tags File
// @Param   uuid     query    string     true        "uuid"
// @Param	userId	query	string	true	"user id"
// @Param	Range header string false  "range, such as：bytes=200-1000"
// @Produce application/octet-stream
// @Failure 404 {object} proto.ErrMess
// @Failure 416 {object} proto.ErrMess "Range Not Satisfiable"
// @Failure 500 {object} proto.ErrMess
// @Success 200 {file}  formData "file content"
// @Success 206 {file}  formData "Partial Content"
// @Router /space/v1/api/file/download [GET]
func DownloadFile(c *gin.Context) {
	ctx := bpctx.NewCtx(c)
	uuid := c.Query("uuid")
	range_ := c.GetHeader("Range")
	ctx.LogD().Str("uuid", uuid).Str("range", range_).Msg("param")
	fileInfo, err := dbutils.GetFileInfoWithUid(ctx.GetUserId(), uuid)
	if err != nil {
		ctx.LogE().Msg(err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, proto.ErrMess{Code: proto.CodeFileNotExist, Message: "File not found"})
		}
		return
	}
	objectKey := fileInfo.BETag
	var part *proto.Part
	if len(range_) > 0 {
		part, err = DecodeRange(range_)
		ctx.LogD().Str("range", range_).Interface("part", part).Msg("range")
		if err != nil {
			c.JSON(http.StatusRequestedRangeNotSatisfiable, proto.ErrMess{Code: proto.CodeParamErr, Message: err.Error()})
			return
		} else if part.End >= fileInfo.Size {
			c.JSON(http.StatusRequestedRangeNotSatisfiable, proto.ErrMess{Code: proto.CodeParamErr, Message: "RangeNotSatisfiable"})
			return
		}
		if part.End == -1 {
			part.End = fileInfo.Size - 1
		}
	}

	r, err := stor.Get(env.NORMAL_BUCKET, objectKey, part)
	if err != nil {
		c.String(http.StatusNotFound, "%s-%v", objectKey, err)
		return
	}
	defer r.Close()

	extraHeaders := map[string]string{
		"Content-Disposition": fmt.Sprintf(`attachment; filename="%s"; filename*=UTF-8''%s`,
			url.QueryEscape(fileInfo.Name),
			url.QueryEscape(fileInfo.Name)),
	}

	if part == nil {
		c.DataFromReader(200, fileInfo.Size, fileInfo.Mime, r, extraHeaders)

	} else {
		c.DataFromReader(206, part.Len(), fileInfo.Mime, r, extraHeaders)
	}
}

// @Summary Get thumbnail
// @Description Get thumbnail
// @Tags File
// @Param   uuid     query    string     true        "uuid"
// @Param	userId	query	string	true	"user id"
// @Failure 404 {object} proto.ErrMess ""
// @Failure 400 {object} proto.ErrMess "param error"
// @Failure 500 {object} proto.ErrMess
// @Success 200 {file}  formData "file content"
// @Router /space/v1/api/file/thumb [GET]
func GetThumb(c *gin.Context) {
	ctx := bpctx.NewCtx(c)
	uuid := c.Query("uuid")
	fileInfo, err := dbutils.GetFileInfoWithUid(ctx.GetUserId(), uuid)
	if err != nil {
		ctx.LogE().Msg(err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, proto.ErrMess{Code: proto.CodeFileNotExist, Message: "File not found"})
		}
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf(`inline; filename="%s"; filename*=UTF-8''%s`,
		url.QueryEscape(fileInfo.Name+"-thumbnail.jpg"), url.QueryEscape(fileInfo.Name+"-thumbnail.jpg"),
	))
	c.Header("ETag", fileInfo.BETag)
	previewStore := storage.NewPreview()
	path, err := previewStore.GetThumbnailPath(fileInfo.BETag)
	if err != nil {
		ctx.LogE().Err(err).Msg("get thumb error")
		c.JSON(http.StatusNotFound, proto.ErrMess{Code: proto.CodeFileNotExist, Message: "File not found"})
		return
	}

	c.File(path)
}

// @Summary  Get compressed graph
// @Description Get compressed graph
// @Tags File
// @Param   uuid     query    string     true        "uuid"
// @Param	userId	query	string	true	"user id"
// @Failure 404 {object} proto.ErrMess ""
// @Failure 500 {object} proto.ErrMess
// @Success 200 {file}  formData "file content"
// @Router /space/v1/api/file/compressed [GET]
func GetCompressed(c *gin.Context) {
	ctx := bpctx.NewCtx(c)
	uuid := c.Query("uuid")
	fileInfo, err := dbutils.GetFileInfoWithUid(ctx.GetUserId(), uuid)
	if err != nil {
		ctx.LogW().Msg(err.Error())
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, proto.ErrMess{Code: proto.CodeFileNotExist, Message: "File not found"})
		}
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf(`inline; filename="%s"; filename*=UTF-8''%s`,
		url.QueryEscape(fileInfo.Name+"-preview.jpg"), url.QueryEscape(fileInfo.Name+"-preview.jpg"),
	))
	c.Header("ETag", fileInfo.BETag)
	previewStore := storage.NewPreview()
	path, err := previewStore.GetCompressedImgPath(fileInfo.BETag)
	if err != nil {
		c.JSON(http.StatusNotFound, proto.ErrMess{Code: proto.CodeFileNotExist, Message: "File not found"})
		return
	}
	c.File(path)

}


