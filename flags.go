// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"github.com/spf13/cobra"
)

var quiet bool

func addQuietFlag(c *cobra.Command) {
	c.Flags().BoolVarP(&quiet, "quiet", "q", false,
		"do not display progress and result of operation")
}

var pkgPath string

func addPkgPathFlag(c *cobra.Command) {
	c.Flags().StringVarP(&pkgPath, "pkgpath", "", "",
		"the list of directories where to search for packages")
}

var workspaceDir string

func addWorkspaceDirFlag(c *cobra.Command) {
	c.Flags().StringVarP(&workspaceDir, "workspacedir", "", ".",
		"pathname of the workspace directory")
}

var defaultMakeTarget string

func addDefaultMakeTargetFlag(c *cobra.Command) {
	c.Flags().StringVarP(&defaultMakeTarget, "maketarget", "", "help",
		"default makefile target")
}

var installDir string

func addInstallDirFlag(c *cobra.Command) {
	c.Flags().StringVarP(&installDir, "installdir", "", "",
		"target directory for 'make install'")
}

var docDir string

func addDocDirFlag(c *cobra.Command) {
	c.Flags().StringVarP(&docDir, "docdir", "", "",
		"installation directory for documentation")
}
