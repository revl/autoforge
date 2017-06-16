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
// Subst returns the number of substitution values.
func (v *verbatim) subst(paramName string, paramValue interface{}) int {
	if textValue, ok := paramValue.(string); ok {
		v.text = strings.Replace(v.text, "{"+paramName+"}",
			textValue, -1)
	} else if arrayValue, ok := paramValue.([]string); !ok {
		v.text = strings.Replace(v.text, "{"+paramName+"}",
			fmt.Sprint(paramValue), -1)
	} else if pos := strings.Index(v.text, "{"+paramName+"}"); pos >= 0 {
		v.next = &array{paramName, arrayValue,
			verbatim{v.text[pos+len(paramName)+2:],
				v.next}}
		v.text = v.text[:pos]
		return len(arrayValue)
	}

	return 1
}

// ExpandPathnameTemplate takes a pathname template and subsitutes
// template parameter names with their values. Parameter values can be
// either strings or slices of strings. Each template value that is a
// slice of strings multiplies the number of output strings by the number
// of strings in the slice.
func expandPathnameTemplate(pathname string,
	params templateParams) []fileParams {
	root := verbatim{pathname, nil}

	resultSize := 1

	for paramName, paramValue := range params {
		resultSize *= root.subst(paramName, paramValue)

		for n := root.next; n != nil; n = n.continuation.next {
			resultSize *= n.continuation.subst(paramName, paramValue)
		}
	}

	result := make([]fileParams, resultSize)

	for i := 0; i < resultSize; i++ {
		result[i].filename = root.text
		copyOfParams := templateParams{}
		for paramName, paramValue := range params {
			copyOfParams[paramName] = paramValue
		}
		result[i].params = copyOfParams
	}

	if resultSize == 0 {
		return result
	}

	sliceSize := resultSize

	for a := root.next; a != nil; a = a.continuation.next {
		numberOfSlices := resultSize / sliceSize
		numberOfValues := len(a.paramValues)
		sliceSize /= numberOfValues

		filenameIndex := 0
		verbatimText := a.continuation.text
		for i := 0; i < numberOfSlices; i++ {
			for j := 0; j < numberOfValues; j++ {
				paramValue := a.paramValues[j]
				filenameFragment := paramValue + verbatimText
				for k := 0; k < sliceSize; k++ {
					result[filenameIndex].filename += filenameFragment
					result[filenameIndex].params[a.paramName] = paramValue
					filenameIndex++
				}
			}
		}
	}

	return result
}
