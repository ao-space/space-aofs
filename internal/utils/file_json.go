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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func WriteJsonToFile(file string, obj interface{}) error {

	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	dir, _ := filepath.Split(file)
	os.MkdirAll(dir, os.ModePerm)

	if f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, os.ModePerm); err != nil {
		return err
	} else {
		if n, err := f.Write(data); err != nil {
			return err
		} else if n != len(data) {
			return fmt.Errorf("The length of data written is insufficient.")
		}
		f.Sync()
		f.Close()
		return nil
	}
}

func ReadJsonFromFile(file string, obj interface{}) error {
	if data, err := os.ReadFile(file); err != nil {
		return err
	} else {
		if err := json.Unmarshal(data, obj); err != nil {
			return err
		} else {
			return nil
		}
	}
}
