// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package logger

import (
	"context"

	"github.com/sirupsen/logrus"
)

type key struct{} // logger key

// WithContext returns a new context with the provided logger
func WithContext(ctx context.Context, logger *logrus.Entry) context.Context {
	return context.WithValue(ctx, key{}, logger)
}

// FromContext retrieves the current logger from the context
func FromContext(ctx context.Context) *logrus.Entry {
	if ctx == nil {
		return logrus.NewEntry(logrus.StandardLogger())
	}

	v := ctx.Value(key{})
	if logger, ok := v.(*logrus.Entry); ok {
		return logger
	}

	return logrus.NewEntry(logrus.StandardLogger())
}
