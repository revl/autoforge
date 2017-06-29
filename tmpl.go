// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type templateParams map[string]interface{}

type fileParams struct {
	filename string
	params   templateParams
}

type pathnameTemplateMultiplier struct {
	paramName    string
	paramValues  []string
	continuation pathnameTemplateText
}

type pathnameTemplateText struct {
	text string
	next *pathnameTemplateMultiplier
}

// Subst updates the 'pathnameTemplateText' receiver by replacing all
// instances of 'name' surrounded by braces with 'value', which can be
// either a string or a slice of strings.  In the latter case, the text
// in the receiver structure gets truncated by the substitution and the
// receiver structure gets extended by a new pathnameTemplateMultiplier
// structure.  Subst returns the number of substitution values.
func (t *pathnameTemplateText) subst(name string, value interface{}) int {
	if textValue, ok := value.(string); ok {
		t.text = strings.Replace(t.text, "{"+name+"}",
			textValue, -1)
	} else if arrayValue, ok := value.([]string); !ok {
		t.text = strings.Replace(t.text, "{"+name+"}",
			fmt.Sprint(value), -1)
	} else if pos := strings.Index(t.text, "{"+name+"}"); pos >= 0 {
		t.next = &pathnameTemplateMultiplier{name, arrayValue,
			pathnameTemplateText{t.text[pos+len(name)+2:], t.next}}
		t.text = t.text[:pos]
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
	root := pathnameTemplateText{pathname, nil}

	resultSize := 1

	for name, value := range params {
		resultSize *= root.subst(name, value)

		for n := root.next; n != nil; n = n.continuation.next {
			resultSize *= n.continuation.subst(name, value)
		}
	}

	result := make([]fileParams, resultSize)

	for i := 0; i < resultSize; i++ {
		result[i].filename = root.text
		copyOfParams := templateParams{}
		for name, value := range params {
			copyOfParams[name] = value
		}
		result[i].params = copyOfParams
	}

	if resultSize == 0 {
		return result
	}

	sliceSize := 1

	for a := root.next; a != nil; a = a.continuation.next {
		numberOfValues := len(a.paramValues)

		continuationText := a.continuation.text

		for i := 0; i < resultSize; {
			for j := 0; j < numberOfValues; j++ {
				value := a.paramValues[j]
				filenameFragment := value + continuationText
				for k := 0; k < sliceSize; k++ {
					result[i].filename += filenameFragment
					result[i].params[a.paramName] = value
					i++
				}
			}
		}

		sliceSize *= numberOfValues
	}

	return result
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func getTemplateWalkFunc(templateDir, projectDir string,
	params templateParams) filepath.WalkFunc {
	return func(templateFile string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		t, err := template.ParseFiles(templateFile)

		if err != nil {
			return err
		}

		if !strings.HasPrefix(templateFile, templateDir) {
			panic(templateFile + " does not start with " +
				templateDir)
		}

		for _, fp := range expandPathnameTemplate(
			templateFile[len(templateDir):], params) {

			projectFile := projectDir + fp.filename

			if err = os.MkdirAll(filepath.Dir(projectFile),
				os.ModePerm); err != nil {
				return err
			}

			f, err := os.OpenFile(projectFile,
				os.O_WRONLY|os.O_CREATE, info.Mode())
			if err != nil {
				return err
			}

			w := bufio.NewWriter(f)

			if err = t.Execute(w, fp.params); err != nil {
				f.Close()
				os.Remove(projectFile)
				return err
			}

			must(w.Flush())
			must(f.Close())
		}

		return nil
	}
}

func generateBuildFilesFromProjectTemplate(templateDir,
	projectDir string, params templateParams) error {

	templateDir = filepath.Clean(templateDir) + string(filepath.Separator)
	projectDir = filepath.Clean(projectDir) + string(filepath.Separator)

	if err := filepath.Walk(templateDir, getTemplateWalkFunc(templateDir,
		projectDir, params)); err != nil {
		return err
	}

	return nil
}
