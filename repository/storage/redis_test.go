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

package storage

import (
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

func TestPushMsg(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		fmt.Println(err)
		return
	}

	InitClient(fmt.Sprintf("%v:%v", mr.Host(), mr.Port()), "", 0)

	PushMsg(map[string]interface{}{"key": "0e01ce6a1829f9a01be7fdd4671fa85280", "bucket": "eulixspace-files"}, "delete")

	time.Sleep(time.Second)
	se, err := mr.Stream("fileChangelogs")
	if err != nil || len(se[0].Values) != 6 {
		t.Fatalf("failed to get msg  from redis. err=%v, values=%v", err, se[0].Values)
	}
}
