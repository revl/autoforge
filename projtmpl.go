// Copyright (C) 2017, 2018 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type fileProcessor func(sourcePathname, relativePathname string,
	info os.FileInfo) error

// processAllFiles calls the processFile() function for every file in
// sourceDir. All hidden files and all files in hidden subdirectories
// as well as package definition files are skipped.
func processAllFiles(sourceDir string, processFile fileProcessor) error {

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

// directoryTree represents a directory structure.
// The 'entries' map contains directory entries.
// If an entry name resolves into nil, it's a file,
// otherwise, it's a subtree.
type directoryTree struct {
	entries map[string]*directoryTree
}

// newDirectoryTree creates a new directory tree
// consisting of an empty root directory.
func newDirectoryTree() *directoryTree {
	return &directoryTree{make(map[string]*directoryTree)}
}

// addFile adds the specified pathname to the directory tree.
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

// hasFile() can be used to check for whether the specified
// file is in the directory tree.
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

// subtree returns a pointer to the branch of the directory tree
// rooted at the specified pathname.
func (dirTree *directoryTree) subtree(pathname string) *directoryTree {
	node := dirTree

	for _, pathComponent := range strings.Split(pathname, "/") {
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

func makePackageDirTree(pd *packageDefinition) (*directoryTree, error) {
	dirTree := newDirectoryTree()

	addFileToDirTree := func(sourcePathname, relativePathname string,
		sourceFileInfo os.FileInfo) error {
		dirTree.addFile(relativePathname)
		return nil
	}

	err := processAllFiles(pd.packageDir(), addFileToDirTree)

	return dirTree, err
}

func pathnamesNotInDirTree(pathnameTemplate string, params templateParams,
	dirTree *directoryTree) []outputFileParams {
	var fileParams []outputFileParams
	for _, fp := range expandPathnameTemplate(pathnameTemplate, params) {
		if !dirTree.hasFile(fp.filename) {
			fileParams = append(fileParams, fp)
		}
	}
	return fileParams
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
	pd *packageDefinition) (bool, error) {

	dirTree, err := makePackageDirTree(pd)
	if err != nil {
		return false, err
	}

	changesMade := false

	for _, fileInfo := range append(t, commonTemplateFiles...) {

		var fileParams []outputFileParams

		for _, fp := range expandPathnameTemplate(fileInfo.pathname,
			pd.params) {
			fileParams = append(fileParams, fp)
		}

		if len(fileParams) == 0 {
			continue
		}

		filesUpdated, err := generateFilesFromProjectFileTemplate(
			fileInfo.pathname, fileInfo.contents,
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

func (pd *packageDefinition) getPackageGeneratorFunc() (
	func() (bool, error), error) {

	switch pd.packageType {
	case "app", "application":
		return func() (bool, error) {
			return generateBuildFilesFromEmbeddedTemplate(
				appTemplate, pd)
		}, nil

	case "lib", "library":
		return func() (bool, error) {
			return generateBuildFilesFromEmbeddedTemplate(
				libTemplate, pd)
		}, nil

	default:
		return nil, errors.New(pd.PackageName +
			": unknown package type '" + pd.packageType + "'")
	}
}
