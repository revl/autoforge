// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"fmt"
	"strings"
)

type templateParams map[string]interface{}

type fileParams struct {
	filename string
	params   templateParams
}

type array struct {
	paramName    string
	paramValues  []string
	continuation verbatim
}

type verbatim struct {
	text string
	next *array
}

// Subst updates the 'verbatim' receiver by replacing all instances
// of paramName surrounded by braces with paramValue, which can be
// either a string or a slice of strings. In the latter case, the
// text in the receiver structure gets truncated by the substitution
// and the receiver structure gets extended by a new array structure.
func (v *verbatim) subst(paramName string, paramValue interface{}) {
	textValue, ok := paramValue.(string)

	if ok {
		v.text = strings.Replace(v.text, "{"+paramName+"}", textValue, -1)
	} else {
		arrayValue, ok := paramValue.([]string)

		if ok {
			pos := strings.Index(v.text, "{"+paramName+"}")

			if pos >= 0 {
				v.next = &array{paramName, arrayValue,
					verbatim{v.text[len(paramName)+2:], v.next}}
				v.text = v.text[:pos]
			}
		} else {
			v.text = strings.Replace(v.text, "{"+paramName+"}",
				fmt.Sprint(paramValue), -1)
		}
	}
}

// ExpandPathnameTemplate takes a pathname template and subsitutes
// template parameter names with their values. Parameter values can be
// either strings or slices of strings. Each template value that is a
// slice of strings multiplies the number of output strings by the number
// of strings in the slice.
func expandPathnameTemplate(pathname string, params templateParams) []fileParams {
	root := verbatim{pathname, nil}

	for paramName, paramValue := range params {
		root.subst(paramName, paramValue)

		for node := root.next; node != nil; node = node.continuation.next {
			node.continuation.subst(paramName, paramValue)
		}

		fmt.Println(paramName, paramValue)
		strings.Index(pathname, "{"+paramName+"}")
	}

	if root.next == nil {
		return []fileParams{{root.text, params}}
	}

	fmt.Println(root)

	return []fileParams{{pathname, params}}
}
