// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"testing"
)

func runTemplateFunctionTest(t *testing.T, funcName, arg, expected string) {

	result := funcMap[funcName].(func(string) string)(arg)

	if result != expected {
		t.Error("Error: \"" + result + "\" != \"" + expected + "\"")
	}
}

func TestTemplateFunctions(t *testing.T) {
	runTemplateFunctionTest(t, "VarName", "C++11", "Cxx11")
	runTemplateFunctionTest(t, "VarName", "one-half", "one_half")

	runTemplateFunctionTest(t, "VarNameUC", "C++11", "CXX11")
	runTemplateFunctionTest(t, "VarNameUC",
		"cross-country", "CROSS_COUNTRY")

	runTemplateFunctionTest(t, "LibName", "libc++11", "libc++11")
	runTemplateFunctionTest(t, "LibName", "dash-dot.", "dash-dot.")
}
