// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

type fileParams struct {
	filename string
	params   interface{}
}

func expandPathnameTemplate(pathname string, params interface{}) []fileParams {
	return []fileParams{{pathname, params}}
}
