// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

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

		if err = generateFilesFromFileTemplate(projectDir,
			relativePathname, templateContents,
			sourceFileInfo.Mode(),
			pd.params, sourceFiles); err != nil {
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
	pathname string
	mode     os.FileMode
	contents []byte
}

// EmbeddedTemplate defines a build-in project template.
type embeddedProjectTemplate []embeddedTemplateFile

// GenerateBuildFilesFromEmbeddedTemplate generates project build
// files from a built-in template pointed to by the 't' parameter.
func generateBuildFilesFromEmbeddedTemplate(t *embeddedProjectTemplate,
	projectDir string, pd *packageDefinition) error {

	sourceFiles, err := linkFilesFromSourceDir(pd, projectDir)
	if err != nil {
		return err
	}

	for _, fileInfo := range *t {
		if _, exists := sourceFiles[fileInfo.pathname]; exists {
			continue
		}

		if err := generateFilesFromFileTemplate(projectDir,
			fileInfo.pathname, fileInfo.contents, fileInfo.mode,
			pd.params, sourceFiles); err != nil {
			return err
		}
	}
	return nil
}
