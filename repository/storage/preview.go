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
	"aofs/internal/env"
	"path/filepath"
)

const compressFileName = "preview.jpg"
const pdfFileName = "preview.pdf"
const thumbFileName = "thumbnail.jpg"

type PreviewStore struct {
	Mdisk  MultiDiskStorager
	Bucket string
}



func NewPreview() *PreviewStore {
	return &PreviewStore{
		Mdisk:  GetStor(),
		Bucket: env.DERIVE_BUCKET,
	}
}

// GetPreviewDir get the object's preview file basic dir path
func (s *PreviewStore) GetPreviewDir(key string) (string, error) {
	diskPath, err := s.Mdisk.GetDiskPathByBEtag(key)
	if err != nil {
		return "", err
	}
	h1 := key[:2]
	h2 := key[2:4]
	dir := filepath.Join(diskPath, s.Bucket, h1, h2, key)
	return dir, nil
}

func (s *PreviewStore) GetPreviewPdfPath(key string) (string, error) {
	baseDir, err := s.GetPreviewDir(key)
	if err != nil {
		return "", err
	}
	path := filepath.Join(baseDir, pdfFileName)
	return path, nil
}

func (s *PreviewStore) GetThumbnailPath(key string) (string, error) {
	baseDir, err := s.GetPreviewDir(key)
	if err != nil {
		return "", err
	}
	path := filepath.Join(baseDir, thumbFileName)
	return path, nil
}

func (s *PreviewStore) GetCompressedImgPath(key string) (string, error) {
	baseDir, err := s.GetPreviewDir(key)
	if err != nil {
		return "", err
	}
	path := filepath.Join(baseDir, compressFileName)
	return path, nil
}


