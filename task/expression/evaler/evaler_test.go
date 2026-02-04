// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package evaler

import (
	"encoding/json"
	"testing"

	"github.com/drone/go-task/task/common"
	"github.com/google/go-cmp/cmp"
)

func TestEval(t *testing.T) {
	var jsondata = []byte(`{
		"static": "value",
		"array": [
			{"static": "value"},
			{"token":  "${{secrets.c94f469b-d84e-4489-9f10-b6b38a7e6023}}"}
		],
		"token": "${{secrets.c94f469b-d84e-4489-9f10-b6b38a7e6023}}"
	}`)

	var data = []*common.Secret{{ID: "c94f469b-d84e-4489-9f10-b6b38a7e6023", Value: "9f105c56f29e4489"}}

	input := map[string]any{}
	err := json.Unmarshal(jsondata, &input)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	Eval(input, data)

	got, want := input, map[string]any{
		"static": "value",
		"array": []any{
			map[string]any{"static": "value"},
			map[string]any{"token": "9f105c56f29e4489"},
		},
		"token": "9f105c56f29e4489",
	}
	if diff := cmp.Diff(got, want); len(diff) != 0 {
		t.Error("Unexpected input expansion")
		t.Log(diff)
	}
}
