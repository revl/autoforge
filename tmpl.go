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

var funcMap = template.FuncMap{
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

		projectFile := filepath.Join(projectDir, fp.filename)

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

type fileProcessor func(sourcePathname, relativePathname string,
	sourceFileInfo os.FileInfo) error

// GetFileGenerationWalkFunc returns a walker function for use with
// filepath.Walk(). The walker function skips directories and package
// definition files.
func getFileGenerationWalkFunc(sourceDir, targetDir string,
	processFile fileProcessor) filepath.WalkFunc {

	sourceDir = filepath.Clean(sourceDir) + string(filepath.Separator)

	return func(sourcePathname string, sourceFileInfo os.FileInfo,
		err error) error {
		if err != nil {
			return err
		}

		if sourceFileInfo.IsDir() {
			return nil
		}

		// Panic if filepath.Walk() does not behave as expected.
		if !strings.HasPrefix(sourcePathname, sourceDir) {
			panic(sourcePathname + " does not start with " +
				sourceDir)
		}

		// Relative pathname of the source file in the source
		// directory (and the target file in the target directory).
		relativePathname := sourcePathname[len(sourceDir):]

		// Ignore the package definition file.
		if relativePathname == packageDefinitionFilename {
			return nil
		}

		return processFile(sourcePathname, relativePathname,
			sourceFileInfo)
	}
}

type filesFromSourceDir map[string]struct{}

func linkFilesFromSourceDir(pd *packageDefinition,
	projectDir string) (filesFromSourceDir, error) {
	sourceFiles := make(filesFromSourceDir)
	sourceDir := filepath.Dir(pd.pathname)

	linkFile := func(sourcePathname, relativePathname string,
		sourceFileInfo os.FileInfo) error {
		sourceFiles[relativePathname] = struct{}{}
		targetPathname := filepath.Join(projectDir, relativePathname)
		targetFileInfo, err := os.Lstat(targetPathname)
		if err == nil {
			if (targetFileInfo.Mode() & os.ModeSymlink) != 0 {
				originalLink, err := os.Readlink(targetPathname)

				if err != nil {
					return err
				}

				if originalLink == sourcePathname {
					return nil
				}
			}

			if err = os.Remove(targetPathname); err != nil {
				return err
			}
		}

		fmt.Println("L", targetPathname)

		if err = os.MkdirAll(filepath.Dir(targetPathname),
			os.ModePerm); err != nil {
			return err
		}

		return os.Symlink(sourcePathname, targetPathname)
	}

	err := filepath.Walk(sourceDir,
		getFileGenerationWalkFunc(sourceDir, projectDir, linkFile))

	return sourceFiles, err
}

// For each source file in 'templateDir', generateBuildFilesFromProjectTemplate
// generates an output file with the same relative pathname inside 'projectDir'.
func generateBuildFilesFromProjectTemplate(templateDir,
	projectDir string, pd *packageDefinition) error {

	sourceFiles, err := linkFilesFromSourceDir(pd, projectDir)
	if err != nil {
		return err
	}

	generateFile := func(sourcePathname, relativePathname string,
		sourceFileInfo os.FileInfo) error {
		if _, sourceFile := sourceFiles[relativePathname]; sourceFile {
			return nil
		}

		// Read the contents of the template file. Cannot use
		// template.ParseFiles() because a Funcs() call must be
		// made between New() and Parse().
		templateContents, err := ioutil.ReadFile(sourcePathname)
		if err != nil {
			return err
		}

		if err = generateFileFromTemplate(projectDir, relativePathname,
			templateContents, sourceFileInfo.Mode(),
			pd.params); err != nil {
			return err
		}

		return nil

	}

	return filepath.Walk(templateDir,
		getFileGenerationWalkFunc(templateDir,
			projectDir, generateFile))
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
	projectDir string, pd *packageDefinition) error {

	sourceFiles, err := linkFilesFromSourceDir(pd, projectDir)
	if err != nil {
		return err
	}

	for pathname, fileInfo := range *template {
		if _, sourceFile := sourceFiles[pathname]; sourceFile {
			continue
		}

		if err := generateFileFromTemplate(projectDir, pathname,
			fileInfo.contents, fileInfo.mode,
			pd.params); err != nil {
			return err
		}
	}
	return nil
}
