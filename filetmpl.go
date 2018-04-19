// Copyright (C) 2017, 2018 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

type templateParams map[string]interface{}

func matchAny(pathname string, patterns []string) bool {
	for _, pattern := range patterns {
		if match, _ := filepath.Match(pattern,
			filepath.Base(pathname)); match {
			return true
		}
	}
	return false
}

func filterPathnames(pathnames, patterns []string, invert bool) []string {
	var filtered []string

	for _, pathname := range pathnames {
		if matchAny(pathname, patterns) != invert {
			filtered = append(filtered, pathname)
		}
	}

	return filtered
}

var commonFuncMap = template.FuncMap{
	"VarName": func(arg string) string {
		return strings.Map(func(r rune) rune {
			if r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' ||
				r >= '0' && r <= '9' {
				return r
			} else if r == '+' {
				return 'x'
			}
			return '_'
		}, arg)
	},
	"VarNameUC": func(arg string) string {
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
	},
	"LibName": func(arg string) string {
		return strings.Map(func(r rune) rune {
			if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' ||
				r >= '0' && r <= '9' ||
				r == '+' || r == '-' || r == '.' {
				return r
			}
			return '_'
		}, arg)
	},
	"TrimExt": func(filename string) string {
		return filename[:len(filename)-len(filepath.Ext(filename))]
	},
	"StringList": func(elem ...string) []string {
		return elem
	},
	"Select": func(pathnames, patterns []string) []string {
		return filterPathnames(pathnames, patterns, false)
	},
	"Exclude": func(pathnames, patterns []string) []string {
		return filterPathnames(pathnames, patterns, true)
	},
	"Comment": func(text string) string {
		var result string

		for _, line := range strings.Split(
			strings.TrimSpace(text), "\n") {
			if line = strings.TrimSpace(line); line != "" {
				result += "# "
				result += line
				result += "\n"
			} else {
				result += "#\n"
			}
		}

		return result + "#\n"
	},
}

type filenameAndContents struct {
	filename string
	contents []byte
}

func parseAndExecuteTemplate(templateName string, templateContents []byte,
	funcMap template.FuncMap, associatedTemplates map[string]string,
	params templateParams) ([]filenameAndContents, error) {

	// Parse the template file. The parsed template will be
	// reused multiple times if expandPathnameTemplate()
	// returns more than one pathname expansion.
	t := template.New(filepath.Base(templateName))
	t.Funcs(commonFuncMap)

	t.Funcs(funcMap)

	for name, text := range associatedTemplates {
		template.Must(t.New(name).Parse(text))
	}

	if _, err := t.Parse(string(templateContents)); err != nil {
		return nil, err
	}

	var result []filenameAndContents

	for _, fp := range expandPathnameTemplate(templateName, params) {
		buffer := bytes.NewBufferString("")

		if err := t.Execute(buffer, fp.params); err != nil {
			return nil, err
		}

		result = append(result, filenameAndContents{
			fp.filename, buffer.Bytes()})
	}

	return result, nil
}

var templateErrorMarker = "AFTMPLERR"

func executePackageFileTemplate(templateName string,
	templateContents []byte, pd *packageDefinition,
	dirTree *directoryTree) ([]filenameAndContents, error) {

	funcMap := template.FuncMap{
		"Error": func(errorMessage string) (string, error) {
			return "", errors.New(templateErrorMarker +
				pd.PackageName + ": " + errorMessage)
		},
		"Dir": func(root string) []string {
			return dirTree.subtree(root).list()
		}}

	return parseAndExecuteTemplate(templateName, templateContents,
		funcMap, commonDefinitions, pd.params)
}

func writeGeneratedFiles(targetDir string, outputFiles []filenameAndContents,
	templateFileMode os.FileMode) (bool, error) {
	targetDir, err := relativeToCwd(targetDir)
	if err != nil {
		return false, err
	}

	changesMade := false
	for _, outputFile := range outputFiles {
		mode := "R"

		projectFile := path.Join(targetDir, outputFile.filename)

		existingFileInfo, err := os.Lstat(projectFile)
		if err != nil {
			if os.IsNotExist(err) {
				if err := os.MkdirAll(filepath.Dir(projectFile),
					os.ModePerm); err != nil {
					return false, err
				}

				mode = "A"
			}
		} else if (existingFileInfo.Mode() & os.ModeSymlink) == 0 {
			oldContents, err := ioutil.ReadFile(projectFile)
			if err == nil {
				if bytes.Compare(oldContents,
					outputFile.contents) == 0 {
					continue
				}
				mode = "U"
			}
		}

		fmt.Println(mode, projectFile)
		if mode == "R" {
			if err = os.Remove(projectFile); err != nil {
				return false, err
			}
		}

		changesMade = true

		if err = ioutil.WriteFile(projectFile, outputFile.contents,
			templateFileMode); err != nil {
			return false, err
		}
	}

	return changesMade, nil
}

func generateFilesFromProjectFileTemplate(projectDir, templateName string,
	templateContents []byte, templateFileMode os.FileMode,
	pd *packageDefinition, dirTree *directoryTree) (bool, error) {

	outputFiles, err := executePackageFileTemplate(templateName,
		templateContents, pd, dirTree)

	if err != nil {
		if err, ok := err.(template.ExecError); ok {
			splitMessage := strings.SplitN(err.Error(),
				templateErrorMarker, 2)

			return false,
				errors.New(splitMessage[len(splitMessage)-1])
		}

		return false, err
	}

	return writeGeneratedFiles(projectDir, outputFiles, templateFileMode)
}
