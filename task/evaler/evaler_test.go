// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package evaler

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestEval(t *testing.T) {
	var jsondata = []byte(`{
		"static": "value",
		"array": [
			{"static": "value"},
			{"token":  "${{secrets.c94f469b-d84e-4489-9f10-b6b38a7e6023}}"},
			{"base64EncodedToken":  "${{getAsBase64(${{secrets.c94f469b-d84e-4489-9f10-b6b38a7e6023}})}}"}
		],
		"token": "${{secrets.c94f469b-d84e-4489-9f10-b6b38a7e6023}}",
                "tokenBase64Encoded": "this is my token: ${{getAsBase64(${{secrets.c94f469b-d84e-4489-9f10-b6b38a7e6023}})}}",
                "mapBase64Encoded": "${{getAsBase64({'key1':'value1','key2':'value2'})}}",
                "emptyBase64Encoded": "${{getAsBase64()}}"
	}`)

	var data = map[string]string{
		"c94f469b-d84e-4489-9f10-b6b38a7e6023": "9f105c56f29e4489",
	}

	input := map[string]any{}
	json.Unmarshal(jsondata, &input)

	Eval(input, data)

	got, want := input, map[string]any{
		"static": "value",
		"array": []any{
			map[string]any{"static": "value"},
			map[string]any{"token": "9f105c56f29e4489"},
			map[string]any{"base64EncodedToken": "OWYxMDVjNTZmMjllNDQ4OQ=="},
		},
		"token":              "9f105c56f29e4489",
		"tokenBase64Encoded": "this is my token: OWYxMDVjNTZmMjllNDQ4OQ==",
		"mapBase64Encoded":   "eydrZXkxJzondmFsdWUxJywna2V5Mic6J3ZhbHVlMid9",
		"emptyBase64Encoded": "",
	}
	if diff := cmp.Diff(got, want); len(diff) != 0 {
		t.Error("Unexpected input expansion")
		t.Log(diff)
	}
}

// func TestExpand(t *testing.T) {
// 	var jsondata = []byte(`{
// 		"static": "value",
// 		"array": [
// 			{"static": "value"}
// 		],
// 		"token": {
// 			"$secret": {
// 				"$id": "c94f469b-d84e-4489-9f10-b6b38a7e6023",
// 				"$path": "data.token"
// 			}
// 		}
// 	}`)

// 	var data = map[string][]byte{
// 		"c94f469b-d84e-4489-9f10-b6b38a7e6023": []byte(`{"data": {"token":"9f105c56f29e4489"}}`),
// 		"5c56f29e-969c-46dc-a150-8c86a2cc297d": []byte(`{}}`),
// 	}

// 	input := map[string]any{}
// 	json.Unmarshal(jsondata, &input)

// 	outputs := Expand(input, data)

// 	{
// 		got, want := outputs, []string{"9f105c56f29e4489"}
// 		if diff := cmp.Diff(got, want); len(diff) != 0 {
// 			t.Error("Unexpected secret output")
// 			t.Log(diff)
// 		}
// 	}

// 	{
// 		got, want := input, map[string]any{
// 			"static": "value",
// 			"array": []any{
// 				map[string]any{"static": "value"},
// 			},
// 			"token": "9f105c56f29e4489",
// 		}
// 		if diff := cmp.Diff(got, want); len(diff) != 0 {
// 			t.Error("Unexpected input expansion")
// 			t.Log(diff)
// 		}
// 	}
// }

// func TestExpand_NotFound(t *testing.T) {
// 	var jsondata = []byte(`{
// 		"static": "value",
// 		"token1": {
// 			"$secret": {
// 				"$id": "c94f469b-d84e-4489-9f10-b6b38a7e6023",
// 				"$path": "data.token1"
// 			}
// 		},
// 		"token2": {
// 			"$secret": {
// 				"$id": "5c56f29e-969c-46dc-a150-8c86a2cc297d",
// 				"$path": "data.token2"
// 			}
// 		}
// 	}`)

// 	var data = map[string][]byte{
// 		"c94f469b-d84e-4489-9f10-b6b38a7e6023": []byte(`{"data": {}}`),
// 		"5c56f29e-969c-46dc-a150-8c86a2cc297d": []byte(`{}}`),
// 	}

// 	input := map[string]any{}
// 	json.Unmarshal(jsondata, &input)
// 	Expand(input, data)

// 	got, want := input, map[string]any{
// 		"static": "value",
// 		"token1": "",
// 		"token2": "",
// 	}
// 	if diff := cmp.Diff(got, want); len(diff) != 0 {
// 		t.Error("Unexpected input expansion")
// 		t.Log(diff)
// 	}
// }
