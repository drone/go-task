// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package common

// Secret stores the value of a secret variable.
type Secret struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}
