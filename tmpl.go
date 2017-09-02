// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
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

// ExpandPathnameTemplate takes a pathname template and substitutes
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

func varName(arg string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' ||
			r >= '0' && r <= '9' {
			return r
		} else if r == '+' {
			return 'x'
		}
		return '_'
	}, arg)
}

func varNameUC(arg string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' {
			return r - 'a' + 'A'
		} else if r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			return r
		} else if r == '+' {
			return 'X'
		}
		return '_'
	}, arg)
}

func libName(arg string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' ||
			r >= '0' && r <= '9' ||
			r == '+' || r == '-' || r == '.' {
			return r
		}
		return '_'
	}, arg)
}

var funcMap template.FuncMap = template.FuncMap{
	"VarName":   varName,
	"VarNameUC": varNameUC,
	"LibName":   libName,
}

func generateFileFromTemplate(projectDir, templatePathname string,
	templateContents []byte, templateFileMode os.FileMode,
	params templateParams) error {
	// Parse the template file. The parsed template will be
	// reused multiple times if expandPathnameTemplate()
	// returns more than one pathname expansion.
	t, err := template.New(filepath.Base(templatePathname)).Funcs(
		funcMap).Parse(string(templateContents))
	if err != nil {
		return err
	}

	for _, fp := range expandPathnameTemplate(templatePathname, params) {

		projectFile := projectDir + fp.filename

		if err = os.MkdirAll(filepath.Dir(projectFile),
			os.ModePerm); err != nil {
			return err
		}

		buffer := bytes.NewBufferString("")

		if err = t.Execute(buffer, fp.params); err != nil {
			return err
		}

		newContents := buffer.Bytes()

		oldContents, err := ioutil.ReadFile(projectFile)

		if os.IsNotExist(err) {
			fmt.Println("A", projectFile)
		} else if bytes.Compare(oldContents, newContents) != 0 {
			fmt.Println("U", projectFile)
		} else {
			return nil
		}

		if err = ioutil.WriteFile(projectFile, newContents,
			templateFileMode); err != nil {
			return err
		}
	}

	return nil
}

// GetTemplateWalkFunc returns a walker function for use with filepath.Walk().
// The returned function interprets each file it visits as a 'text/template'
// file and generates a new file with the same relative pathname in the output
// directory 'projectDir'.
func getTemplateWalkFunc(templateDir, projectDir string,
	params templateParams) filepath.WalkFunc {
	return func(templateFile string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Panic if filepath.Walk() does not behave as expected.
		if !strings.HasPrefix(templateFile, templateDir) {
			panic(templateFile + " does not start with " +
				templateDir)
		}

		// Read the contents of the template file. Cannot use
		// template.ParseFiles() because a Funcs() call must be
		// made between New() and Parse().
		templateContents, err := ioutil.ReadFile(templateFile)
		if err != nil {
			return err
		}

		// Pathname of the template file relative to the
		// template directory.
		templateFile = templateFile[len(templateDir):]

		// Ignore package definition file for the template.
		if templateFile == filepath.Base(templateDir)+".yaml" {
			return nil
		}

		if err = generateFileFromTemplate(projectDir, templateFile,
			templateContents, info.Mode(), params); err != nil {
			return err
		}

		return nil
	}
}

// For each source file in 'templateDir', generateBuildFilesFromProjectTemplate
// generates an output file with the same relative pathname inside 'projectDir'.
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

// EmbeddedTemplateFile defines the file mode and the contents
// of a single file that is a part of an embedded project template.
type embeddedTemplateFile struct {
	mode     os.FileMode
	contents []byte
}

// EmbeddedTemplate defines a build-in project template.
type embeddedTemplate map[string]embeddedTemplateFile

// GenerateBuildFilesFromEmbeddedTemplate generates project build
// files from a built-in template pointed to by the 'template' parameter.
func generateBuildFilesFromEmbeddedTemplate(template *embeddedTemplate,
	projectDir string, params templateParams) error {
	for pathname, file_info := range *template {
		if err := generateFileFromTemplate(projectDir, pathname,
			file_info.contents, file_info.mode,
			params); err != nil {
			return err
		}
	}
	return nil
}
