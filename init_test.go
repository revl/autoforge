// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import "testing"

func Test(t *testing.T) {
	if loadInitParams().OutputDir == "" {
		t.Errorf("Oops")
	}
}
