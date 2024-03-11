// Copyright 2024 Harness Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package logger

// func TestContext(t *testing.T) {
// 	entry := Discard()

// 	ctx := WithContext(context.Background(), entry)
// 	got := FromContext(ctx)

// 	if got != entry {
// 		t.Errorf("Expected Logger from context")
// 	}
// }

// func TestEmptyContext(t *testing.T) {
// 	got := FromContext(context.Background())
// 	if got == nil {
// 		t.Errorf("Expected Logger from context")
// 	}
// 	if _, ok := got.(*discard); !ok {
// 		t.Errorf("Expected discard Logger from context")
// 	}
// }
