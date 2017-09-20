// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"github.com/spf13/cobra"
)

var flags = struct {
	quiet             bool
	pkgPath           string
	workspaceDir      string
	defaultMakeTarget string
	installDir        string
	docDir            string
}{}

func addQuietFlag(c *cobra.Command) {
	c.Flags().BoolVarP(&flags.quiet, "quiet", "q", false,
		"do not display progress and result of operation")
}

func addPkgPathFlag(c *cobra.Command) {
	c.Flags().StringVarP(&flags.pkgPath, "pkgpath", "", "",
		"the list of directories where to search for packages")
}

func addWorkspaceDirFlag(c *cobra.Command) {
	c.Flags().StringVarP(&flags.workspaceDir, "workspacedir", "", ".",
		"pathname of the workspace directory")
}

func addDefaultMakeTargetFlag(c *cobra.Command) {
	c.Flags().StringVarP(&flags.defaultMakeTarget, "maketarget", "", "help",
		"default makefile target")
}

func addInstallDirFlag(c *cobra.Command) {
	c.Flags().StringVarP(&flags.installDir, "installdir", "", "",
		"target directory for 'make install'")
}

func addDocDirFlag(c *cobra.Command) {
	c.Flags().StringVarP(&flags.docDir, "docdir", "", "",
		"installation directory for documentation")
}
