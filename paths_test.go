// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"strings"
	"testing"
)

func substTestCase(t *testing.T, node *verbatim, expected string) {
	var result string

	for {
		result += node.text
		if node.next == nil {
			break
		}

		result += "[" + strings.Join(node.next.paramValues, ", ") + "]"

		node = &node.next.continuation
	}

	if result != expected {
		t.Error("Error: \"" + result + "\" != \"" + expected + "\"")
	}
}

func TestSubst(t *testing.T) {
	v := verbatim{"{Greetings}, {Who}!", nil}
	v.subst("Who", "Human")
	v.subst("Greetings", []string{"Hello", "Hi"})

	substTestCase(t, &v, "[Hello, Hi], Human!")

	v = verbatim{"{What}, {What} {Where}", nil}
	v.subst("What", "Mirror")
	v.subst("Where", "on the Wall")

	substTestCase(t, &v, "Mirror, Mirror on the Wall")
}
