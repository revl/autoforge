// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"reflect"
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

func runExpandPathnameTemplateTest(t *testing.T,
	pathname string, params map[string]interface{}, expected []fileParams) {

	result := expandPathnameTemplate(pathname, params)
	if !reflect.DeepEqual(result, expected) {
		t.Error("Result", result,
			"and expected result", expected, "are not equal")
	}
}

func TestExpandPathnameTemplate(t *testing.T) {
	params2x3 := map[string]interface{}{
		"name": []string{"foo", "bar"},
		"ext":  []string{"js", "go", "rs"}}

	result2x3 := []fileParams{
		{"foo.js", templateParams{"name": "foo", "ext": "js"}},
		{"bar.js", templateParams{"name": "bar", "ext": "js"}},
		{"foo.go", templateParams{"name": "foo", "ext": "go"}},
		{"bar.go", templateParams{"name": "bar", "ext": "go"}},
		{"foo.rs", templateParams{"name": "foo", "ext": "rs"}},
		{"bar.rs", templateParams{"name": "bar", "ext": "rs"}}}

	runExpandPathnameTemplateTest(t, "{name}.{ext}", params2x3, result2x3)

	params4x2x3 := map[string]interface{}{
		"dir":  []string{"A", "B", "C", "D"},
		"name": []string{"1", "2"},
		"ext":  []string{"a", "b", "c"}}

	result4x2x3 := []fileParams{
		{"A/1.a", templateParams{"dir": "A", "name": "1", "ext": "a"}},
		{"B/1.a", templateParams{"dir": "B", "name": "1", "ext": "a"}},
		{"C/1.a", templateParams{"dir": "C", "name": "1", "ext": "a"}},
		{"D/1.a", templateParams{"dir": "D", "name": "1", "ext": "a"}},
		{"A/2.a", templateParams{"dir": "A", "name": "2", "ext": "a"}},
		{"B/2.a", templateParams{"dir": "B", "name": "2", "ext": "a"}},
		{"C/2.a", templateParams{"dir": "C", "name": "2", "ext": "a"}},
		{"D/2.a", templateParams{"dir": "D", "name": "2", "ext": "a"}},
		{"A/1.b", templateParams{"dir": "A", "name": "1", "ext": "b"}},
		{"B/1.b", templateParams{"dir": "B", "name": "1", "ext": "b"}},
		{"C/1.b", templateParams{"dir": "C", "name": "1", "ext": "b"}},
		{"D/1.b", templateParams{"dir": "D", "name": "1", "ext": "b"}},
		{"A/2.b", templateParams{"dir": "A", "name": "2", "ext": "b"}},
		{"B/2.b", templateParams{"dir": "B", "name": "2", "ext": "b"}},
		{"C/2.b", templateParams{"dir": "C", "name": "2", "ext": "b"}},
		{"D/2.b", templateParams{"dir": "D", "name": "2", "ext": "b"}},
		{"A/1.c", templateParams{"dir": "A", "name": "1", "ext": "c"}},
		{"B/1.c", templateParams{"dir": "B", "name": "1", "ext": "c"}},
		{"C/1.c", templateParams{"dir": "C", "name": "1", "ext": "c"}},
		{"D/1.c", templateParams{"dir": "D", "name": "1", "ext": "c"}},
		{"A/2.c", templateParams{"dir": "A", "name": "2", "ext": "c"}},
		{"B/2.c", templateParams{"dir": "B", "name": "2", "ext": "c"}},
		{"C/2.c", templateParams{"dir": "C", "name": "2", "ext": "c"}},
		{"D/2.c", templateParams{"dir": "D", "name": "2", "ext": "c"}}}

	runExpandPathnameTemplateTest(t, "{dir}/{name}.{ext}",
		params4x2x3, result4x2x3)

	paramsNil := map[string]interface{}{
		"nil":      []string{},
		"noeffect": []string{"value"}}

	resultNil := []fileParams{}

	runExpandPathnameTemplateTest(t, "{nil}/{noeffect}",
		paramsNil, resultNil)
}
