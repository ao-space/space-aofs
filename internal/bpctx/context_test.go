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

package bpctx

import (
	"testing"

	"github.com/rs/zerolog"
)

func TestLog(t *testing.T) {
	// c := &gin.Context{}
	ctx := NewCtx(nil)
	ctx.GetRawLogger().Level(zerolog.DebugLevel)
	defer ctx.LogI("title", "req test")
	ctx.LogD().Str("debugD", "test logD").Msg("logD")
	ctx.GetRawLogger().Debug().Str("debubk", "test debug").Msg("xx")
	//ctx.SendErr(proto.CodeFailedToConnect, fmt.Errorf("test error log"))
}
