// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package task

import (
	"encoding/json"
	"fmt"
)

// Response is a response interface.
type Response interface {
	// Body gets the response body.
	Body() []byte

	// Error gets the response error.
	Error() error
}

// Respond creates a response.
func Respond(v any) Response {
	var b []byte
	switch t := v.(type) {
	case []byte:
		b = t
	case string:
		b = []byte(t)
	case nil:
		break
	default:
		var err error
		if b, err = json.Marshal(t); err != nil {
			return Error(err)
		}
	}
	return &Result{
		Data: b,
	}
}

// Error creates an error response.
func Error(err error) Response {
	return &Result{
		Err: err,
	}
}

// Errorf creates an error response.
func Errorf(format string, a ...any) Response {
	return &Result{
		Err: fmt.Errorf(format, a...),
	}
}

//
//
//

// Result provides task results.
type Result struct {
	Err     error
	Secrets map[string]string
	Outputs map[string]string
	Data    []byte
}

// Body gets the response body.
func (r *Result) Body() []byte {
	return r.Data
}

// Error gets the response error.
func (r *Result) Error() error {
	return r.Err
}
