// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package task

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/drone/go-task/task/common"
)

var noContext = context.Background()

func TestRouter(t *testing.T) {
	router := NewRouter()
	router.RegisterFunc("ping", func(_ context.Context, req *Request) Response {
		return Respond("pong")
	})
	router.NotFoundFunc(func(_ context.Context, req *Request) Response {
		t.Fail()
		return nil
	})

	res := router.Handle(noContext, &Request{
		Task: &Task{
			Type: "ping",
		},
	})

	got := res.Body()
	want := []byte("pong")
	if !bytes.Equal(got, want) {
		t.Errorf("Want response body %s, got %s", want, got)
	}
}

func TestRouterErr(t *testing.T) {
	router := NewRouter()
	router.RegisterFunc("ping", func(_ context.Context, req *Request) Response {
		return Errorf("ping error")
	})

	res := router.Handle(noContext, &Request{
		Task: &Task{
			Type: "ping",
		},
	})

	got := res.Error()
	want := "ping error"
	if got.Error() != want {
		t.Errorf("Want response error %s, got %s", want, got)
	}
}

func TestRouterErr_NotFound(t *testing.T) {
	router := NewRouter()

	res := router.Handle(noContext, &Request{
		Task: &Task{
			Type: "ping",
		},
	})

	got := res.Error()
	want := "handler not found"
	if got.Error() != want {
		t.Errorf("Want response error %s, got %s", want, got)
	}
}

func TestRouter_NotFound(t *testing.T) {
	router := NewRouter()
	router.NotFoundFunc(func(_ context.Context, req *Request) Response {
		return Errorf("custom not found error")
	})

	res := router.Handle(noContext, &Request{
		Task: &Task{
			Type: "ping",
		},
	})

	got := res.Error()
	want := "custom not found error"
	if got.Error() != want {
		t.Errorf("Want response error %s, got %s", want, got)
	}
}

func TestMiddleware(t *testing.T) {
	var visited1, visited2, visited3 bool
	r := NewRouter()
	r.Use(
		func(next Handler) Handler {
			return HandlerFunc(
				func(ctx context.Context, req *Request) Response {
					visited1 = true
					return next.Handle(ctx, req)
				},
			)
		},
	)
	r.Use(
		func(next Handler) Handler {
			return HandlerFunc(
				func(ctx context.Context, req *Request) Response {
					visited2 = true
					return next.Handle(ctx, req)
				},
			)
		},
	)
	r.RegisterFunc("test", func(_ context.Context, req *Request) Response {
		visited3 = true
		return nil
	})

	req := &Request{Task: &Task{Type: "test"}}

	if err := r.Handle(context.Background(), req); err != nil {
		t.Error(err)
	}
	if !visited1 {
		t.Errorf("Expect middleware[1] invoked")
	}
	if !visited2 {
		t.Errorf("Expect middleware[2] invoked")
	}
	if !visited3 {
		t.Errorf("Expect handler invoked")
	}
}

func TestMiddleware_Break(t *testing.T) {
	var visited bool
	r := NewRouter()
	r.Use(
		func(next Handler) Handler {
			return HandlerFunc(
				func(ctx context.Context, req *Request) Response {
					visited = true
					return nil // break chain
				},
			)
		},
	)
	r.Use(
		func(next Handler) Handler {
			return HandlerFunc(
				func(ctx context.Context, req *Request) Response {
					t.Fail()
					return next.Handle(ctx, req)
				},
			)
		},
	)
	r.RegisterFunc("test", func(_ context.Context, req *Request) Response {
		t.Fail()
		return nil
	})

	req := &Request{Task: &Task{Type: "test"}}

	if err := r.Handle(context.Background(), req); err != nil {
		t.Error(err)
	}

	if !visited {
		t.Fail()
	}
}

func TestMiddlewareErr(t *testing.T) {
	r := NewRouter()
	r.Use(
		func(next Handler) Handler {
			return HandlerFunc(
				func(_ context.Context, req *Request) Response {
					return Error(fmt.Errorf("test error"))
				},
			)
		},
	)
	r.Use(
		func(next Handler) Handler {
			return HandlerFunc(
				func(ctx context.Context, req *Request) Response {
					t.Fail()
					return next.Handle(ctx, req)
				},
			)
		},
	)
	r.RegisterFunc("test", func(_ context.Context, req *Request) Response {
		t.Fail()
		return nil
	})

	req := &Request{Task: &Task{Type: "test"}}

	if err := r.Handle(context.Background(), req); err == nil {
		t.Errorf("Expect middleware error")
	}
}

func TestResolveSecrets(t *testing.T) {
	router := NewRouter()
	router.RegisterFunc("secret_task", func(_ context.Context, req *Request) Response {
		return Respond(&common.Secret{Value: "mySecret"})
	})

	got, err := router.ResolveSecrets(noContext, []*Task{{ID: "secret_task_id", Type: "secret_task"}})
	if err != nil {
		t.Errorf("error when resolving secrets: %s", err)
	}

	want := []*common.Secret{{ID: "secret_task_id", Value: "mySecret"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Want resolved secrets %v, got %v", want, got)
	}
}

func TestResolveExpressions(t *testing.T) {
	router := NewRouter()
	secrets := []*common.Secret{{ID: "secret_task_id", Value: "mySecret"}}
	taskData, _ := json.Marshal(map[string]string{"taskKey": "this is my secret: ${{secrets.secret_task_id}}"})

	resolvedTaskData, _, err := router.ResolveExpressions(noContext, secrets, taskData)
	if err != nil {
		t.Errorf("error when resolving expressions: %s", err)
	}

	got := make(map[string]string)
	err = json.Unmarshal(resolvedTaskData, &got)
	if err != nil {
		t.Errorf("error converting resolved task data to map: %s", err)
	}

	want := map[string]string{"taskKey": "this is my secret: mySecret"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Want resolved task data with expressions %v, got %v", want, got)
	}
}
