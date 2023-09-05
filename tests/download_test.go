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

package tests

import (
	"aofs/internal/proto"
	"aofs/routers/api"
	"reflect"
	"testing"
)

func testDownloadAll(t *testing.T) {
	t.Run("testDecodeRange", testDecodeRange)
}

func testDecodeRange(t *testing.T) {
	type args struct {
		rangestr string
	}
	tests := []struct {
		name    string
		args    args
		want    *proto.Part
		wantErr bool
	}{
		// TODO: Add test cases.
		{"", args{"bytes=0-"}, &proto.Part{Start: 0, End: -1}, false},
		{"", args{"bytes=0-5"}, &proto.Part{Start: 0, End: 5}, false},
		{"", args{"bytes=-5"}, &proto.Part{Start: 0, End: 5}, false},
		{"", args{"bytes=-5,"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := api.DecodeRange(tt.args.rangestr)
			if (err != nil) != tt.wantErr {
				t.Errorf("decodeRange() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("decodeRange() = %v, want %v", got, tt.want)
			}
		})
	}
}
