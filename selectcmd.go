// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

func generateAndBootstrapPackage(pkgSelection []string) error {
	packageIndex, err := buildPackageIndex()
	if err != nil {
		return err
	}

	pkgRootDir := filepath.Join(getWorkspaceDir(), "packages")

	type packageAndGenerator struct {
		pd         *packageDefinition
		packageDir string
		generator  func() error
	}

	var packagesAndGenerators []packageAndGenerator

	for _, packageName := range pkgSelection {
		pd, ok := packageIndex.packageByName[packageName]
		if !ok {
			return errors.New("no such package: " + packageName)
		}

		packageDir := filepath.Join(pkgRootDir, pd.packageName)

		generator, err := pd.getPackageGeneratorFunc(packageDir)
		if err != nil {
			return err
		}

		packagesAndGenerators = append(packagesAndGenerators,
			packageAndGenerator{pd, packageDir, generator})
	}

	for _, pg := range packagesAndGenerators {
		// Generate autoconf and automake sources for the package.
		if err = pg.generator(); err != nil {
			return err
		}

		// Bootstrap the package if the 'configure' script
		// does not exist.
		_, err = os.Lstat(filepath.Join(pg.packageDir, "configure"))
		if os.IsNotExist(err) {
			cmd := exec.Command("./autogen.sh")
			cmd.Dir = pg.packageDir
			if err = cmd.Run(); err != nil {
				return err
			}
		}
	}

	return nil
}

// SelectCmd represents the select command
var selectCmd = &cobra.Command{
	Use:   "select package_range...",
	Short: "Choose one or more packages to work on",
	Args:  cobra.MinimumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		if err := generateAndBootstrapPackage(args); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(selectCmd)

	selectCmd.Flags().SortFlags = false
	addQuietFlag(selectCmd)
	addWorkspaceDirFlag(selectCmd)
	addPkgPathFlag(selectCmd)
}
