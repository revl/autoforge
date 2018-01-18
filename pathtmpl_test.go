// Copyright (C) 2017, 2018 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"reflect"
	"strings"
	"testing"
)

func substTestCase(t *testing.T, node *pathnameTemplateText, expected string) {
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
	v := pathnameTemplateText{"{Greetings}, {Who}!", nil}
	v.subst("Who", "Human")
	v.subst("Greetings", []string{"Hello", "Hi"})

	substTestCase(t, &v, "[Hello, Hi], Human!")

	v = pathnameTemplateText{"{What}, {What} {Where}", nil}
	v.subst("What", "Mirror")
	v.subst("Where", "on the Wall")

	substTestCase(t, &v, "Mirror, Mirror on the Wall")
}

func runExpandPathnameTemplateTest(t *testing.T,
	pathname string, params map[string]interface{},
	expected []outputFileParams) {

	result := expandPathnameTemplate(pathname, params)
	if !reflect.DeepEqual(result, expected) {
		t.Error("Result", result,
			"and expected result", expected, "are not equal")
	}
}

func TestExpandPathnameTemplate2x3(t *testing.T) {
	params2x3 := map[string]interface{}{
		"name": []string{"foo", "bar"},
		"ext":  []string{"js", "go", "rs"}}

	var result2x3 []outputFileParams

	for _, ext := range params2x3["ext"].([]string) {
		for _, name := range params2x3["name"].([]string) {
			filename := name + "." + ext

			result2x3 = append(result2x3,
				outputFileParams{filename,
					templateParams{
						"name":     name,
						"ext":      ext,
						"filename": filename,
						"dirname":  "."}})
		}
	}

	runExpandPathnameTemplateTest(t, "{name}.{ext}", params2x3, result2x3)
}

func TestExpandPathnameTemplate4x2x3(t *testing.T) {
	params4x2x3 := map[string]interface{}{
		"dir":  []string{"A", "B", "C", "D"},
		"name": []string{"1", "2"},
		"ext":  []string{"a", "b", "c"}}

	var result4x2x3 []outputFileParams

	for _, ext := range params4x2x3["ext"].([]string) {
		for _, name := range params4x2x3["name"].([]string) {
			for _, dir := range params4x2x3["dir"].([]string) {
				filename := dir + "/" + name + "." + ext

				result4x2x3 = append(result4x2x3,
					outputFileParams{filename,
						templateParams{
							"dir":      dir,
							"name":     name,
							"ext":      ext,
							"filename": filename,
							"dirname":  dir}})
			}
		}
	}

	runExpandPathnameTemplateTest(t, "{dir}/{name}.{ext}",
		params4x2x3, result4x2x3)
}

func TestExpandPathnameTemplateNoFiles(t *testing.T) {
	paramsNil := map[string]interface{}{
		"nil":      []string{},
		"noeffect": []string{"value"}}

	resultNil := []outputFileParams{}

	runExpandPathnameTemplateTest(t, "{nil}/{noeffect}",
		paramsNil, resultNil)
}
