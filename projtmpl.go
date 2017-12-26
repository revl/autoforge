// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type fileProcessor func(sourcePathname, relativePathname string,
	info os.FileInfo) error

// ProcessAllFiles calls the processFile() function for every file in
// sourceDir. All hidden files and all files in hidden subdirectories
// as well as package definition files are skipped.
func processAllFiles(sourceDir, targetDir string,
	processFile fileProcessor) error {

	sourceDir = filepath.Clean(sourceDir)
	sourceDirWithSlash := sourceDir + string(filepath.Separator)

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

	err := processAllFiles(sourceDir, projectDir, linkFile)

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

		return generateFilesFromFileTemplate(projectDir,
			relativePathname, templateContents,
			sourceFileInfo.Mode(),
			pd, sourceFiles)
	}

	return processAllFiles(templateDir, projectDir, generateFile)
}

// EmbeddedTemplateFile defines the file mode and the contents
// of a single file that is a part of an embedded project template.
type embeddedTemplateFile struct {
	pathname string
	mode     os.FileMode
	contents []byte
}

// GenerateBuildFilesFromEmbeddedTemplate generates project build
// files from a built-in template pointed to by the 't' parameter.
func generateBuildFilesFromEmbeddedTemplate(t []embeddedTemplateFile,
	projectDir string, pd *packageDefinition) error {

	sourceFiles, err := linkFilesFromSourceDir(pd, projectDir)
	if err != nil {
		return err
	}

	for _, fileInfo := range append(t, commonTemplateFiles...) {
		if _, exists := sourceFiles[fileInfo.pathname]; exists {
			continue
		}

		if err := generateFilesFromFileTemplate(projectDir,
			fileInfo.pathname, fileInfo.contents, fileInfo.mode,
			pd, sourceFiles); err != nil {
			return err
		}
	}

	return nil
}

func (pd *packageDefinition) getPackageGeneratorFunc(
	packageDir string) (func() error, error) {
	switch pd.packageType {
	case "app", "application":
		return func() error {
			return generateBuildFilesFromEmbeddedTemplate(
				appTemplate, packageDir, pd)
		}, nil

	case "lib", "library":
		return func() error {
			return generateBuildFilesFromEmbeddedTemplate(
				libTemplate, packageDir, pd)
		}, nil

	default:
		return nil, errors.New(pd.packageName +
			": unknown package type '" + pd.packageType + "'")
	}
}
