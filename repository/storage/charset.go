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
	"bufio"
	"aofs/internal/env"
	"os"

	"github.com/saintfish/chardet"
)

func GetCharset(betag string) string {
	filePath, err := GetStor().GetFileAbsPath(env.NORMAL_BUCKET, betag)
	if err != nil {
		return ""
	}
	f, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	buf := make([]byte, 1024)
	_, err = reader.Read(buf)
	if err != nil {
		return ""
	}
	detector := chardet.NewTextDetector()
	charset, err := detector.DetectBest(buf)

	return charset.Charset
}
