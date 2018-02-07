// Copyright (C) 2017, 2018 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"os"
	"path"
	"path/filepath"
)

var pkgDirName = "packages"

func getGeneratedPkgRootDir(privateDir string) string {
	return path.Join(privateDir, pkgDirName)
}

func getBuildDir(privateDir string, wp *workspaceParams) string {
	if wp.BuildDir != "" {
		return wp.BuildDir
	}
	return path.Join(privateDir, "build")
}

func relativeToCwd(absPath string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Rel(cwd, absPath)
}

func absIfNotEmpty(pathname string) (string, error) {
	if pathname == "" {
		return "", nil
	}
	return filepath.Abs(pathname)
}
