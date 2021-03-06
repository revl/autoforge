// Copyright (C) 2017, 2018 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var flags = struct {
	quiet             bool
	pkgPath           string
	workspaceDir      string
	makefile          string
	defaultMakeTarget string
	buildDir          string
	installDir        string
	noBootstrap       bool
}{}

func addQuietFlag(c *cobra.Command) {
	c.Flags().BoolVarP(&flags.quiet, "quiet", "q", false,
		"do not display progress and result of operation")
}

func addPkgPathFlag(c *cobra.Command) {
	c.Flags().StringVar(&flags.pkgPath, "pkgpath", "",
		"the list of directories where to search for packages")
}

func getPkgPathFlag() (string, error) {
	pkgpath := flags.pkgPath
	if pkgpath == "" {
		return "", nil
	}

	var absPaths []string
	for _, colonSeparated := range strings.Split(pkgpath, ":") {
		if colonSeparated != "" {
			absPath, err := filepath.Abs(colonSeparated)
			if err != nil {
				return "", err
			}
			absPaths = append(absPaths, absPath)
		}
	}

	return strings.Join(absPaths, ":"), nil
}

func addWorkspaceDirFlag(c *cobra.Command) {
	c.Flags().StringVar(&flags.workspaceDir, "workspacedir", ".",
		"pathname of the workspace directory")
}

func addMakefileFlag(c *cobra.Command) {
	c.Flags().StringVar(&flags.makefile, "makefile", "",
		"filename of the generated makefile (default \"Makefile\")")
}

const maketargetOption = "maketarget"

func addDefaultMakeTargetFlag(c *cobra.Command) {
	c.Flags().StringVar(&flags.defaultMakeTarget, maketargetOption, "",
		"first target in the makefile (default \"help\")")
}

func addBuildDirFlag(c *cobra.Command) {
	c.Flags().StringVar(&flags.buildDir, "builddir", "",
		"directory for building the packages")
}

func addInstallDirFlag(c *cobra.Command) {
	c.Flags().StringVar(&flags.installDir, "installdir", "",
		"target directory for 'make install'")
}

func addNoBootstrapFlag(c *cobra.Command) {
	c.Flags().BoolVarP(&flags.noBootstrap, "nobootstrap", "", false,
		"do not bootstrap packages ("+conftabFilename+
			" will not be updated)")
}
