// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package task

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"io"

	"github.com/drone/go-task/task/common"
	"github.com/drone/go-task/task/expression"
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
		WithFields(map[string]interface{}{
			"task.id":     req.Task.ID,
			"task.type":   req.Task.Type,
			"task.driver": req.Task.Driver,
		})
	log.Debug("route task")

	// handle each secret sub-task before handling
	// the primary task
	var err error
	req.Secrets, err = h.ResolveSecrets(ctx, req.Tasks)
	if err != nil {
		return Error(err)
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

func (h *Router) ResolveSecrets(ctx context.Context, tasks []*Task) ([]*common.Secret, error) {
	secrets := []*common.Secret{}

	// handle each secret sub-task
	for _, subtask := range tasks {
		subreq := new(Request)
		subreq.Task = subtask
		subreq.Secrets = secrets

		// handle the subtask and get the results.
		res := h.handle(ctx, subreq)

		// immediately exit if the system fails
		// to execute the secret task.
		if err := res.Error(); err != nil {
			return nil, err
		}

		var secretOutputBytes []byte
		if subtask.Driver != "cgi" && subtask.Type != "cgi" {
			// This is not CGI task
			secretOutputBytes = res.Body()
		} else {
			// This is CGI task
			// Decode the response body into a temporary
			// data structure.
			out := new(CGITaskResponse)
			if err := json.Unmarshal(res.Body(), out); err != nil {
				return nil, err
			}

			if decodedBody, err := base64.StdEncoding.DecodeString(out.Body); err != nil {
				return nil, fmt.Errorf("failed to decode plugin response: %s. %s", subtask.ID, err)
			} else {
				if out.StatusCode > 299 {
					// Check whether it's a successful CGI call. Fail the task if it's not, as we can't proceed without the secret.
					return nil, fmt.Errorf("failed to retrieve secret: %s. %s", subtask.ID, decodedBody)
				} else {
					secretOutputBytes = decodedBody
				}
			}
		}

		secretOutput := new(common.Secret)
		if err := json.Unmarshal(secretOutputBytes, secretOutput); err != nil {
			return nil, fmt.Errorf("failed to unmarshal secret: %s. %s", subtask.ID, err)
		}
		secretOutput.ID = subtask.ID

		// add the secret to request
		secrets = append(secrets, secretOutput)
	}
	return secrets, nil
}

func (h *Router) ResolveExpressions(ctx context.Context, secrets []*common.Secret, taskData []byte) ([]byte, error) {
	resolver := expression.New(secrets)
	resolvedTaskData, err := resolver.Resolve(taskData)
	if err != nil {
		return nil, err
	}
	return resolvedTaskData, nil
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

	// evaluate expressions
	var err error
	req.Task.Data, err = h.ResolveExpressions(ctx, req.Secrets, req.Task.Data)
	if err != nil {
		return Error(err)
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
