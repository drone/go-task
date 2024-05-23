// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package task

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/drone/go-task/task/evaler"
	"github.com/drone/go-task/task/logger"
)

// Router routes task execution requests to the
// appropriate handler.
type Router struct {
	middleware []func(Handler) Handler
	handlers   map[string]Handler
	notfound   Handler
}

func NewRouter() *Router {
	return &Router{
		handlers: map[string]Handler{},
	}
}

// Use adds the middleware onto the router stack.
func (h *Router) Use(fn func(Handler) Handler) {
	h.middleware = append(h.middleware, fn)
}

// Register registers the Handler to the router.
func (h *Router) Register(name string, handler Handler) {
	h.handlers[name] = handler
}

// RegisterFunc registers the HandlerFunc to the router.
func (h *Router) RegisterFunc(name string, handler HandlerFunc) {
	h.Register(name, HandlerFunc(handler))
}

// NotFound adds a handler to response whenver a
// route cannot be found.
func (h *Router) NotFound(handler Handler) {
	h.notfound = handler
}

// NotFoundFunc adds a handler to response whenver a
// route cannot be found.
func (h *Router) NotFoundFunc(handler HandlerFunc) {
	h.NotFound(HandlerFunc(handler))
}

// Handle routes the task request to a handler.
func (h *Router) Handle(ctx context.Context, req *Request) Response {
	log := logger.FromContext(ctx).
		With("task.id", req.Task.ID).
		With("task.type", req.Task.Type).
		With("task.driver", req.Task.Driver)

	log.Debug("route task")

	// ensure all required variables are initialized.
	if req.Secrets == nil {
		req.Secrets = map[string]string{}
	}

	// handle each secret sub-task before handling
	// the primary sub-task
	for _, subtask := range req.Tasks {
		subreq := new(Request)
		subreq.Task = subtask
		subreq.Tasks = req.Tasks
		subreq.Secrets = req.Secrets

		// handle the subtask and get the results.
		res := h.handle(ctx, subreq)

		// immediately exit if the system fails
		// to execute the secret task.
		if err := res.Error(); err != nil {
			return res
		}

		// attempt to unmarshal the task response
		// body into the secrets struct.
		out := new(Secret)
		if err := json.Unmarshal(res.Body(), &out); err != nil {
			return Error(err)
		}

		// add the secret to request
		req.Secrets[subtask.ID] = out.Value
	}

	// add the structured logger to the context.
	ctx = logger.WithContext(ctx, log)

	// Discard task logs if a logger is not set.
	// A custom logger can be set by adding a middleware to the router.
	if req.Logger == nil {
		req.Logger = io.Discard
	}

	// handle the primary task
	return h.handle(ctx, req)
}

// handle routes the task request to a handler.
func (h *Router) handle(ctx context.Context, req *Request) Response {
	// extract the task type
	name := req.Task.Type

	// lookup the task handler
	handler, ok := h.handlers[name]
	if !ok {
		// error if no route found
		if h.notfound == nil {
			return Errorf("handler not found")
		}

		// else use the not found handler
		// to handle the task.
		handler = h.notfound
	}

	// TODO(bradrydzewski) move expression eval to middleware?
	if bytes.Contains(req.Task.Data, []byte("${{")) {
		v := map[string]any{}

		// unmarshal the task data into a map
		err := json.Unmarshal(req.Task.Data, &v)
		if err != nil {
			return Error(err)
		}

		// evaluate the expressions
		evaler.Eval(v, req.Secrets)

		// encode the map back to json
		req.Task.Data, err = json.Marshal(v)
		if err != nil {
			return Error(err)
		}
	}

	// execute the handler stack with middleware
	return chain(h.middleware, handler).Handle(ctx, req)
}

// chain builds a Handler composed of an inline
// middleware stack and endpoint handler in the
// order they are passed.
func chain(middleware []func(Handler) Handler, handler Handler) Handler {
	// return ahead of time if there aren't any
	// middleware for the chain
	if len(middleware) == 0 {
		return handler
	}

	// wrap the end handler with the middleware chain
	h := middleware[len(middleware)-1](handler)
	for i := len(middleware) - 2; i >= 0; i-- {
		h = middleware[i](h)
	}

	return h
}
