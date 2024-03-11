// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package logger

import (
	"context"
	"log/slog"
)

type key struct{} // logger key

// WithContext returns a new context with the provided logger.
func WithContext(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, key{}, logger)
}

// FromContext retrieves the current logger from the context.
func FromContext(ctx context.Context) *slog.Logger {
	// if nil, return the defualt logger
	if ctx == nil {
		return slog.Default()
	}
	v := ctx.Value(key{})
	// return the valid logger if returned
	if logger, ok := v.(*slog.Logger); ok {
		return logger
	}
	// else return the default logger
	return slog.Default()
}
