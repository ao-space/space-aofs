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

package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
)

//FormatGinLog 日志格式化
func LoggerHandler() gin.HandlerFunc {

	var statusColor, resetColor, level string
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {

		if param.IsOutputColor() {
			statusColor = param.StatusCodeColor()
			resetColor = param.ResetColor()
		}
		if param.StatusCode >= 300 {
			level = "error"
		} else {
			level = "info"
		}
		return fmt.Sprintf("{\"level\":\"%s\",\"timestamp\":%s,\"client\":\"%s\",\"method\":\"%s\",\"path\":\"%s\",\"request.proto\":\"%s\",\"code\":\"%s %3d %s\",\"latency\":\"%s\"}\n",
			level,
			param.TimeStamp.Format(time.RFC3339),
			param.ClientIP,
			param.Method,
			param.Path,
			param.Request.Proto,
			statusColor, param.StatusCode, resetColor,
			param.Latency,
		)
	})
}
