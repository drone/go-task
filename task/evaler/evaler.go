// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package evaler provides helper functions for evaluating
// expressions in map structures.

package evaler

import (
	"encoding/base64"
	"regexp"
	"strings"
)

// Eval evaluates expressions in the map structure.
func Eval(data map[string]any, secrets map[string]string) {
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
			v = resolveGetAsBase64(v)
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

func resolveSecrets(s string, secrets map[string]string) string {
	// HACK(bradrydzewski) find/replace secrets
	// until we have proper expression support.
	for key, value := range secrets {
		a := "${{secrets." + key + "}}"
		b := value
		s = strings.ReplaceAll(s, a, b)
	}
	return s
}

func resolveGetAsBase64(s string) string {
	// regex to match the pattern ${{getAsBase64(...)}}
	re := regexp.MustCompile(`\$\{\{getAsBase64\((.*?)\)\}\}`)

	// replace the matched strings with their base64-encoded values
	return re.ReplaceAllStringFunc(s, func(match string) string {
		// extract the string wrapped by ${{getAsBase64(...)}}
		submatch := re.FindStringSubmatch(match)
		if len(submatch) > 1 {
			// Encode it to to base64
			return base64.StdEncoding.EncodeToString([]byte(submatch[1]))
		}
		return match
	})
}
