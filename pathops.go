// Copyright (C) 2017, 2018 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"os"
	"path/filepath"
)

var pkgDirName = "packages"

func getGeneratedPkgRootDir(privateDir string) string {
	return privateDir + "/" + pkgDirName
}

func getBuildDir(privateDir string, wp *workspaceParams) string {
	if wp.BuildDir != "" {
		return wp.BuildDir
	}
	return privateDirName + "/build"
}

func relativeToCwd(absPath string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Rel(cwd, absPath)
}
