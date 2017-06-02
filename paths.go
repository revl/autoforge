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

func (self *verbatim) subst(paramName string, paramValue interface{}) {
	textValue, ok := paramValue.(string)
	if ok {
		self.text = strings.Replace(self.text, "{"+paramName+"}", textValue, -1)
	} else {
		arrayValue, ok := paramValue.([]string)
		if ok {
			pos := strings.Index(self.text, "{"+paramName+"}")
			if pos >= 0 {
				self.next = &array{paramName, arrayValue,
					verbatim{self.text[len(paramName)+2:], self.next}}
				self.text = self.text[:pos]
			}
		} else {
			self.text = strings.Replace(self.text, "{"+paramName+"}",
				fmt.Sprint(paramValue), -1)
		}
	}
}

func expandPathnameTemplate(pathname string, params templateParams) []fileParams {
	root := &verbatim{pathname, nil}

	for paramName, paramValue := range params {
		root.subst(paramName, paramValue)

		for node := root.next; node != nil; node = node.continuation.next {
			node.continuation.subst(paramName, paramValue)
		}

		fmt.Println(paramName, paramValue)
		strings.Index(pathname, "{"+paramName+"}")
	}

	fmt.Println(root)

	return []fileParams{{pathname, params}}
}
