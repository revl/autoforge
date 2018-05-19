// Copyright (C) 2017, 2018 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

type fileProcessor func(sourcePathname, relativePathname string,
	info os.FileInfo) error

// processAllFiles calls the processFile() function for every file in
// sourceDir. All hidden files and all files in hidden subdirectories
// as well as package definition files are skipped.
func processAllFiles(sourceDir, targetDir string,
	processFile fileProcessor) error {

	sourceDir = filepath.Clean(sourceDir)
	sourceDirWithSlash := sourceDir + "/"

	return filepath.Walk(sourceDir, func(sourcePathname string,
		info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Ignore the top-level directory (sourceDir itself).
		if len(sourcePathname) <= len(sourceDirWithSlash) {
			return nil
		}

		// Panic if filepath.Walk() does not behave as expected.
		if !strings.HasPrefix(sourcePathname, sourceDirWithSlash) {
			panic(sourcePathname + " does not start with " +
				sourceDirWithSlash)
		}

		// Relative pathname of the source file in the source
		// directory (and the target file in the target directory).
		relativePathname := sourcePathname[len(sourceDirWithSlash):]

		// Ignore hidden files and the package definition file.
		if filepath.Base(relativePathname)[0] == '.' {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		} else if info.IsDir() {
			return nil
		} else if relativePathname == packageDefinitionFilename {
			return nil
		}

		return processFile(sourcePathname, relativePathname, info)
	})
}

type directoryTree struct {
	entries map[string]*directoryTree
}

func newDirectoryTree() *directoryTree {
	return &directoryTree{make(map[string]*directoryTree)}
}

func (dirTree *directoryTree) addFile(filePath string) {
	pathComponents := strings.Split(filePath, "/")

	nComp := len(pathComponents)
	if nComp == 0 {
		return
	}

	node := dirTree

	for _, pathComponent := range pathComponents[:nComp-1] {
		child := node.entries[pathComponent]
		if child == nil {
			child = newDirectoryTree()
			node.entries[pathComponent] = child
		}
		node = child
	}

	node.entries[pathComponents[nComp-1]] = nil
}

func (dirTree *directoryTree) hasFile(filePath string) bool {
	pathComponents := strings.Split(filePath, "/")

	nComp := len(pathComponents)
	if nComp == 0 {
		return false
	}

	node := dirTree

	for _, pathComponent := range pathComponents[:nComp-1] {
		child := node.entries[pathComponent]
		if child == nil {
			return false
		}
		node = child
	}

	entry, entryExists := node.entries[pathComponents[nComp-1]]
	return entryExists && entry == nil
}

func (dirTree *directoryTree) subtree(filePath string) *directoryTree {
	node := dirTree

	for _, pathComponent := range strings.Split(filePath, "/") {
		if pathComponent == "." {
			continue
		}
		child := node.entries[pathComponent]
		if child == nil {
			return nil
		}
		node = child
	}

	return node
}

func listFiles(basePath string, dirTree *directoryTree) []string {
	var list []string

	basePath += "/"

	for entry, child := range dirTree.entries {
		if child == nil {
			list = append(list, basePath+entry)
		} else {
			list = append(list, listFiles(basePath+entry, child)...)
		}
	}

	return list
}

func (dirTree *directoryTree) list() []string {
	var list []string

	for entry, child := range dirTree.entries {
		if child == nil {
			list = append(list, entry)
		} else {
			list = append(list, listFiles(entry, child)...)
		}
	}

	sort.Strings(list)

	return list
}

func linkFilesFromSourceDir(pd *packageDefinition,
	projectDir string) (*directoryTree, bool, error) {
	dirTree := newDirectoryTree()
	sourceDir := filepath.Dir(pd.pathname)
	changesMade := false

	linkFile := func(sourcePathname, relativePathname string,
		sourceFileInfo os.FileInfo) error {
		dirTree.addFile(relativePathname)
		targetPathname := path.Join(projectDir, relativePathname)
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

		changesMade = true

		return os.Symlink(sourcePathname, targetPathname)
	}

	err := processAllFiles(sourceDir, projectDir, linkFile)

	return dirTree, changesMade, err
}

func pathnamesNotInDir(pathnameTemplate string, params templateParams,
	dirTree *directoryTree) []outputFileParams {
	var fileParams []outputFileParams
	for _, fp := range expandPathnameTemplate(pathnameTemplate, params) {
		if !dirTree.hasFile(fp.filename) {
			fileParams = append(fileParams, fp)
		}
	}
	return fileParams
}

// generateBuildFilesFromProjectTemplate generates an output file inside
// 'projectDir' with the same relative pathname as the respective source
// file in 'templateDir'.
func generateBuildFilesFromProjectTemplate(templateDir,
	projectDir string, pd *packageDefinition) (bool, error) {

	dirTree, changesMade, err := linkFilesFromSourceDir(pd, projectDir)
	if err != nil {
		return false, err
	}

	generateFile := func(sourcePathname, relativePathname string,
		sourceFileInfo os.FileInfo) error {
		fileParams := pathnamesNotInDir(relativePathname,
			pd.params, dirTree)

		if len(fileParams) == 0 {
			return nil
		}

		// Read the contents of the template file. Cannot use
		// template.ParseFiles() because a Funcs() call must be
		// made between New() and Parse().
		templateContents, err := ioutil.ReadFile(sourcePathname)
		if err != nil {
			return err
		}

		filesUpdated, err := generateFilesFromProjectFileTemplate(
			projectDir, relativePathname, templateContents,
			sourceFileInfo.Mode(), pd, dirTree, fileParams)
		if err != nil {
			return err
		}
		if filesUpdated {
			changesMade = true
		}
		return nil
	}

	err = processAllFiles(templateDir, projectDir, generateFile)

	return changesMade, err
}

// embeddedTemplateFile defines the file mode and the contents
// of a single file that is a part of an embedded project template.
type embeddedTemplateFile struct {
	pathname string
	mode     os.FileMode
	contents []byte
}

// generateBuildFilesFromEmbeddedTemplate generates project build
// files from a built-in template pointed to by the 't' parameter.
func generateBuildFilesFromEmbeddedTemplate(t []embeddedTemplateFile,
	projectDir string, pd *packageDefinition) (bool, error) {

	dirTree, changesMade, err := linkFilesFromSourceDir(pd, projectDir)
	if err != nil {
		return false, err
	}

	for _, fileInfo := range append(t, commonTemplateFiles...) {
		fileParams := pathnamesNotInDir(fileInfo.pathname,
			pd.params, dirTree)

		if len(fileParams) == 0 {
			continue
		}

		filesUpdated, err := generateFilesFromProjectFileTemplate(
			projectDir, fileInfo.pathname, fileInfo.contents,
			fileInfo.mode, pd, dirTree, fileParams)
		if err != nil {
			return false, err
		}
		if filesUpdated {
			changesMade = true
		}
	}

	return changesMade, nil
}

func (pd *packageDefinition) getPackageGeneratorFunc(
	packageDir string) (func() (bool, error), error) {
	switch pd.packageType {
	case "app", "application":
		return func() (bool, error) {
			return generateBuildFilesFromEmbeddedTemplate(
				appTemplate, packageDir, pd)
		}, nil

	case "lib", "library":
		return func() (bool, error) {
			return generateBuildFilesFromEmbeddedTemplate(
				libTemplate, packageDir, pd)
		}, nil

	default:
		return nil, errors.New(pd.PackageName +
			": unknown package type '" + pd.packageType + "'")
	}
}
