// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package task

import "context"

// A Handler handles task execution.
type Handler interface {
	Handle(context.Context, *Request) Response
}

// The HandlerFunc type is an adapter to allow the use of
// ordinary functions as task handlers.
type HandlerFunc func(context.Context, *Request) Response

// Handle calls f.
func (f HandlerFunc) Handle(ctx context.Context, req *Request) Response {
	return f(ctx, req)
}
