// Copyright (C) 2017, 2018 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

var pkgDirName = "packages"

func getGeneratedPkgRootDir(privateDir string) string {
	return privateDir + "/" + pkgDirName
}

func getBuildDir(privateDir string, wp *workspaceParams) string {
	if flags.buildDir != "" {
		return flags.buildDir
	}
	if wp.BuildDir != "" {
		return wp.BuildDir
	}
	return privateDirName + "/build"
}
