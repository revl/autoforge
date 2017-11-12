// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
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

var funcMap = template.FuncMap{
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
	"StringList": func(elem ...string) []string {
		return elem
	},
	"Select": func(pathnames, patterns []string) []string {
		return filterPathnames(pathnames, patterns, false)
	},
	"Exclude": func(pathnames, patterns []string) []string {
		return filterPathnames(pathnames, patterns, true)
	},
	"Error": func(errorMessage string) (string, error) {
		return "", errors.New(errorMessage)
	},
	"Comment": func(text string) string {
		return strings.Replace(strings.TrimSpace(text),
			"\n", "\n# ", -1)
	},
}

type filenameAndContents struct {
	filename string
	contents []byte
}

func executeFileTemplate(templatePathname string,
	templateContents []byte, params templateParams,
	sourceFiles filesFromSourceDir) ([]filenameAndContents, error) {

	// Parse the template file. The parsed template will be
	// reused multiple times if expandPathnameTemplate()
	// returns more than one pathname expansion.
	t := template.New(filepath.Base(templatePathname))
	t.Funcs(funcMap)

	t.Funcs(template.FuncMap{
		"Dir": func(root string) []string {
			root += string(filepath.Separator)

			var filtered []string

			for sourceFile := range sourceFiles {
				if strings.HasPrefix(sourceFile, root) {
					filtered = append(filtered,
						sourceFile[len(root):])
				}
			}

			sort.Strings(filtered)

			return filtered
		}})

	for name, text := range commonTemplates {
		template.Must(t.New(name).Parse(text))
	}

	if _, err := t.Parse(string(templateContents)); err != nil {
		return nil, err
	}

	var result []filenameAndContents

	for _, fp := range expandPathnameTemplate(templatePathname, params) {
		buffer := bytes.NewBufferString("")

		if err := t.Execute(buffer, fp.params); err != nil {
			return nil, err
		}

		result = append(result, filenameAndContents{
			fp.filename, buffer.Bytes()})
	}

	return result, nil
}

func generateFilesFromFileTemplate(projectDir, templatePathname string,
	templateContents []byte, templateFileMode os.FileMode,
	params templateParams, sourceFiles filesFromSourceDir) error {

	outputFiles, err := executeFileTemplate(templatePathname,
		templateContents, params, sourceFiles)

	if err != nil {
		return err
	}

	for _, outputFile := range outputFiles {
		mode := "R"

		projectFile := filepath.Join(projectDir, outputFile.filename)

		existingFileInfo, err := os.Lstat(projectFile)
		if err != nil {
			if os.IsNotExist(err) {
				if err := os.MkdirAll(filepath.Dir(projectFile),
					os.ModePerm); err != nil {
					return err
				}

				mode = "A"
			}
		} else if (existingFileInfo.Mode() & os.ModeSymlink) == 0 {
			oldContents, err := ioutil.ReadFile(projectFile)
			if err == nil {
				if bytes.Compare(oldContents,
					outputFile.contents) == 0 {
					return nil
				}
				mode = "U"
			}
		}

		fmt.Println(mode, projectFile)
		if mode == "R" {
			if err = os.Remove(projectFile); err != nil {
				return err
			}
		}

		if err = ioutil.WriteFile(projectFile, outputFile.contents,
			templateFileMode); err != nil {
			return err
		}
	}

	return nil
}
