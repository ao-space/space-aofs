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
	"bytes"
	"crypto/rand"
	"io/ioutil"
	"testing"
)

func TestCrypt(t *testing.T) {
	data := make([]byte, 139)
	rand.Read(data)

	r := NewCryptReader(bytes.NewReader(data), "123")
	encData, err := ioutil.ReadAll(r)

	if err != nil {
		t.Fatal(err)
	}
	if len(data) != len(encData) {
		t.Fatal("size err")
	}

	rPlain := NewCryptReader(bytes.NewReader(encData), "123")
	plainData, err := ioutil.ReadAll(rPlain)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != len(plainData) {
		t.Fatal("size err")
	}
	if bytes.Compare(data, plainData) != 0 {
		t.Fatal("decrypt err")
	}

}

func TestCrypt2(t *testing.T) {
	for i := 0; i < 100; i++ {
		data := make([]byte, 139+i)
		rand.Read(data)

		bufer := bytes.NewBuffer(nil)
		w := NewCryptWriter(bufer, "123")

		if n, err := w.Write(data); err != nil || n != len(data) || n != (139+i) {
			t.Fatal(err)
		}
		if len(data) != len(bufer.Bytes()) {
			t.Fatal("size err")
		}

		rPlain := NewCryptReader(bytes.NewReader(bufer.Bytes()), "123")
		plainData, err := ioutil.ReadAll(rPlain)
		if err != nil {
			t.Fatal(err)
		}
		if len(data) != len(plainData) {
			t.Fatal("size err")
		}
		if bytes.Compare(data, plainData) != 0 {
			t.Fatal("decrypt err")
		}

	}

}
