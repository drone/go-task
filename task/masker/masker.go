// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package masker

import (
	"io"
	"strings"
)

// replacer is an io.Writer that finds and masks
// sensitive data.
type replacer struct {
	w io.Writer
	r *strings.Replacer
}

// New returns a masker that wraps io.Writer w.
func New(w io.Writer, secrets []string) io.Writer {
	var oldnew []string
	for _, secret := range secrets {
		for _, part := range strings.Split(secret, "\n") {
			part = strings.TrimSpace(part)

			if len(part) == 0 {
				continue
			}

			masked := "[redacted]"
			oldnew = append(oldnew, part)
			oldnew = append(oldnew, masked)
		}
	}
	if len(oldnew) == 0 {
		return w
	}
	return &replacer{
		w: w,
		r: strings.NewReplacer(oldnew...),
	}
}

// Write writes p to the base writer. The method scans for any
// sensitive data in p and masks before writing.
func (r *replacer) Write(p []byte) (n int, err error) {
	_, err = r.w.Write([]byte(r.r.Replace(string(p))))
	return len(p), err
}

// Slice converts a key value pair of secrets to a slice.
func Slice(in map[string]string) (out []string) {
	for _, v := range in {
		out = append(out, v)
	}
	return out
}
