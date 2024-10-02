// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package task

// Request defines a task request.
type Request struct {
	// Task provides the current task.
	Task *Task `json:"task"`

	// Tasks provides the previous task
	// execution results.
	Tasks []*Task `json:"secrets"`

	// Secrets provides the names and values of secrets
	// that are available to the task execution.
	Secrets map[string]string `json:"-"`

	// Account provides the account identifier.
	Account string `json:"account"`

	// ID provides a unique identifier to track the status of the request.
	ID string `json:"id"`
}
