// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package evaler provides helper functions for evaluating
// expressions in map structures.

package evaler

import (
	"strings"

	"github.com/drone/go-task/task/common"
)

// Eval evaluates expressions in the map structure.
func Eval(data map[string]any, secrets []*common.Secret) {
	var walk func(any) (bool, string)

	// helper function to walk the map and inject
	// secret variables in child keys where the
	// value is a reference to a $secret.
	walk = func(i any) (_ bool, _ string) {
		switch v := i.(type) {
		case string:
			if !strings.Contains(v, "${{") {
				return
			}
			v = resolveSecrets(v, secrets)
			return true, v
		case []any:
			for i := 0; i < len(v); i++ {
				if ok, updated := walk(v[i]); ok {
					v[i] = updated
				}
			}
		case map[string]any:
			for key, value := range v {
				if ok, updated := walk(value); ok {
					v[key] = updated
				}
			}
		}
		return
	}

	// walk the map
	walk(data)
}

func resolveSecrets(s string, secrets []*common.Secret) string {
	// HACK(bradrydzewski) find/replace secrets
	// until we have proper expression support.
	for _, secret := range secrets {
		key := secret.ID
		value := secret.Value
		a := "${{secrets." + key + "}}"
		b := value
		s = strings.ReplaceAll(s, a, b)
	}
	return s
}
