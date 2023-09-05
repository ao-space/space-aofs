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
	"log"
	"syscall"
	"unsafe"
)

func DiskUsage(path string) (disk DiskStatus) {
	kernel32, err := syscall.LoadLibrary("Kernel32.dll")
	if err != nil {
		log.Panic(err)
	}
	defer syscall.FreeLibrary(kernel32)
	GetDiskFreeSpaceEx, err := syscall.GetProcAddress(kernel32, "GetDiskFreeSpaceExW")

	if err != nil {
		log.Panic(err)
	}

	lpFreeBytesAvailable := int64(0)
	lpTotalNumberOfBytes := int64(0)
	lpTotalNumberOfFreeBytes := int64(0)
	_, _, _ = syscall.Syscall6(GetDiskFreeSpaceEx, 4,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("D:"))),
		uintptr(unsafe.Pointer(&lpFreeBytesAvailable)),
		uintptr(unsafe.Pointer(&lpTotalNumberOfBytes)),
		uintptr(unsafe.Pointer(&lpTotalNumberOfFreeBytes)), 0, 0)

	log.Printf("Available  %db", lpFreeBytesAvailable)
	log.Printf("Total      %db", lpTotalNumberOfBytes)
	log.Printf("Free       %db", lpTotalNumberOfFreeBytes)
	disk.Free = uint64(lpTotalNumberOfFreeBytes)
	disk.All = uint64(lpTotalNumberOfBytes)
	disk.Used = uint64(lpTotalNumberOfBytes - lpTotalNumberOfFreeBytes)
	return
}
