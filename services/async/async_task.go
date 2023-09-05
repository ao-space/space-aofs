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

package async

import (
	"aofs/internal/utils"
	"errors"
	"sync"
)

// 定义一个任务的struct，包含任务ID，任务处理总数，已处理数
type AsyncTask struct {
	TaskId     string `json:"taskId"`
	TaskStatus string `json:"taskStatus"`
	Total      int    `json:"total"`
	Processed  int    `json:"processed"`
}

func (a *AsyncTask) UpdateStatus(status string) {
	a.TaskStatus = status
}

func (a *AsyncTask) Init(total int) {
	a.TaskStatus = AsyncTaskStatusInit
	a.Total = total
	a.TaskId = utils.RandomID()[:8]
	a.Processed = 0
}

const (
	AsyncTaskStatusInit       = "init"
	AsyncTaskStatusProcessing = "processing"
	AsyncTaskStatusSuccess    = "success"
	AsyncTaskStatusFailed     = "failed"
)

// 任务初始化
type TaskList struct {
	mu    sync.Mutex
	tasks map[string]*AsyncTask
}

//var tl *TaskList

func NewTaskList() *TaskList {
	return &TaskList{
		tasks: make(map[string]*AsyncTask),
	}
}

func (t *TaskList) Add(task *AsyncTask) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.tasks[task.TaskId] = task
}

func (t *TaskList) Get(taskID string) *AsyncTask {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.tasks[taskID]
}

func (t *TaskList) Remove(taskID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.tasks, taskID)
}

// GetTaskStatus 获取任务状态
func (t *TaskList) GetTaskStatus(taskID string) (*AsyncTask, error) {
	// 获取任务列表
	//taskList := NewTaskList()
	// 根据任务 ID 获取任务
	task := t.Get(taskID)
	// 返回任务和错误
	if task == nil {
		return nil, errors.New("task not found")
	}
	return task, nil
}
