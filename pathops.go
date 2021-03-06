// Copyright (C) 2017, 2018 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"os"
	"path/filepath"
)

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

// relativeIfShorter returns a pathname relative between the first
// and the second pathname arguments under condition that both
// arguments are absolute pathnames and the resulting relative
// pathname is shorter than the second argument.
func relativeIfShorter(basePath, targetPath string) string {
	relPath, err := filepath.Rel(basePath, targetPath)
	if err == nil && len(relPath) < len(targetPath) {
		return relPath
	}
	return targetPath
}
