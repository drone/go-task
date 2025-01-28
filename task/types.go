// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package task

type Task struct {
	// ID provides a unique task identifier.
	ID string `json:"id"`

	// Type provides the task type.
	Type string `json:"type"`

	// Data provides task execution data.
	Data []byte `json:"data"`

	// Driver provides the execution driver used to
	// execute the task.
	Driver string `json:"driver"`

	// Config provides the execution driver configuration.
	Config []byte `json:"config"`

	// Forward provides instructions for forwarding
	// the task to another runner node in the network.
	Forward *Forward `json:"forward"`

	// Logger provides instructions on where to log the output.
	Logger *Logger `json:"logger"`
}

// Forward provides instructions for forward a task
// to another runner node in the network.
type Forward struct {
	Address  string `json:"string"`
	Insecure bool   `json:"insecure"`
	Certs    Certs  `json:"certs"`
}

// Logger provides instructions for logging the output
// of a task execution.
type Logger struct {
	Address           string   `json:"address"`
	Insecure          bool     `json:"insecure"`
	Token             string   `json:"token"`
	Key               string   `json:"key"`
	Account           string   `json:"account"`
	IndirectUpload    bool     `json:"indirect_upload"`
	SkipOpeningStream bool     `json:"skip_opening_stream"`
	Masks             []string `json:"masks"`
}

// Certs provides tls certificates.
type Certs struct {
	Public  []byte `json:"public"`
	Private []byte `json:"private"`
	CA      []byte `json:"ca"`
}

// Repository provides the repository used for
// custom task execution.
type Repository struct {
	Clone    string `json:"clone"`
	Ref      string `json:"ref"`
	Sha      string `json:"sha"`
	Download string `json:"download"`
}

// ExecutableConfig provides the details to download
// a custom binary task executable file, for all
// supported operating systems and architectures
type ExecutableConfig struct {
	Executables []Executable `json:"executables"`
	Name        string       `json:"name"`
	Version     string       `json:"version"`
}

// Executable provides the url to download
// a custom binary task executable file,
// given the operating system and architecture
type Executable struct {
	Arch string `json:"arch"`
	Os   string `json:"os"`
	Url  string `json:"url"`
}

// type State struct {
// 	// ID provides a unique task identifier.
// 	ID string `json:"id"`

// 	// Status provides the task status.
// 	Status Status `json:"status,omitempty"`

// 	// Started provides the task start date.
// 	Started int64 `json:"started,omitempty"`

// 	// Finished provides the task end date.
// 	Finished int64 `json:"finished,omitempty"`
// }

// // Config configures the execution driver.
// type Config struct {
// 	Command []string `json:"command"`
// 	Args    []string `json:"args"`
// 	Envs    []string `json:"envs"`

// 	// Artifact provides instructions for downloading
// 	// and unpacking an artifact used for task exection,
// 	// such as a custom binary, docker image, etc.
// 	Artifact *Artifact `json:"artifact"`
// }

// // Logger provides the logger endpoint details.
// type Logger struct {
// 	Address  string `json:"address"`
// 	Insecure bool   `json:"insecure"`
// 	Token    string `json:"token"`
// }

// type Status string

// const (
// 	StatusUnknown = Status("")
// 	StatusSuccess = Status("success")
// 	StatusFailure = Status("failure")
// 	StatusRunning = Status("running")
// )

// type Driver string

// const (
// 	DriverNone  = Status("")
// 	DriverExec  = Status("exec")
// 	DriverCGI   = Status("cgi")
// 	DriverHTTP  = Status("http")
// 	DriverProxy = Status("proxy")
// )
