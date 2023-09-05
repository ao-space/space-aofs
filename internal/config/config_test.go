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

package config

import (
	"os"
	"testing"
)

func TestReadString(t *testing.T) {
	type args struct {
		key string
		def string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"conf_a", args{"conf_a", "a"}, "a"},
		{"conf_a", args{"conf_a", ""}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReadString(tt.args.key, tt.args.def); got != tt.want {
				t.Errorf("ReadString, actual = %v, want %v", got, tt.want)
			}
		})
	}

	os.Setenv("conf_a", "aenv")
	os.Setenv("CONF_B", "benv")
	os.Setenv("CONF_c", "cenv")
	os.Setenv("CONF_d", "")

	tests2 := []struct {
		name string
		args args
		want string
	}{
		{"conf_a", args{"conf_a", "a"}, "aenv"},
		{"conf_b", args{"conf_b", "b"}, "benv"},
		{"conf_c", args{"conf_c", "c"}, "cenv"},
		{"conf_d", args{"conf_d", "d"}, ""},
	}
	for _, tt := range tests2 {
		{
			t.Run(tt.name, func(t *testing.T) {
				if got := ReadString(tt.args.key, tt.args.def); got != tt.want {
					t.Errorf("ReadString actual = %v, want %v", got, tt.want)
				}
			})
		}
	}

}

func TestReadInt(t *testing.T) {
	type args struct {
		key string
		def int
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{"conf_a", args{"conf_a", 11}, 11},
		{"conf_b", args{"conf_b", 0}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReadInt(tt.args.key, tt.args.def); got != tt.want {
				t.Errorf("ReadString actual = %v, want %v", got, tt.want)
			}
		})
	}

	os.Setenv("conf_a", "aenv")
	os.Setenv("CONF_B", "benv")
	os.Setenv("CONF_c", "cenv")
	os.Setenv("CONF_d", "")

	tests2 := []struct {
		name string
		args args
		want int
	}{
		{"conf_a", args{"conf_a", 1}, 1},
		{"conf_b", args{"conf_b", 2}, 2},
		{"conf_c", args{"conf_c", -3}, -3},
		{"conf_d", args{"conf_d", 4}, 4},
	}
	for _, tt := range tests2 {
		{
			t.Run(tt.name, func(t *testing.T) {
				if got := ReadInt(tt.args.key, tt.args.def); got != tt.want {
					t.Errorf("ReadString: actual = %v, want %v", got, tt.want)
				}
			})
		}
	}

}
