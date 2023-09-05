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

package utils

import (
	"testing"
)

func TestGetMimeTypeByFilename(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{"null", args{""}, "application/octet-stream"},
		{"jpeg", args{"x.jpg"}, "image/jpeg"},
		{"unknown", args{"x.jpgxxxxxx"}, "application/octet-stream"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetMimeTypeByFilename(tt.args.filename); got != tt.want {
				t.Errorf("GetMimeTypeByFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseCategoryByFilename(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{"null", args{""}, "other"},
		{"jpeg", args{"x.jpg"}, "picture"},
		{"jpeg", args{"x.avi"}, "video"},
		{"jpeg", args{"x.pptx"}, "document"},
		{"unknown", args{"x.jpgxxxxxx"}, "other"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseCategoryByFilename(tt.args.filename); got != tt.want {
				t.Errorf("ParseCategoryByFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}
