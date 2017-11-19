// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"errors"
	"log"
	"path/filepath"

	"github.com/spf13/cobra"
)

func generatePackageSources(packages []string) error {
	packageIndex, err := buildPackageIndex()
	if err != nil {
		return err
	}

	buildDir := filepath.Join(getWorkspaceDir(), "build")

	var projectGenerators []func() error

	for _, pkg := range packages {
		pd, ok := packageIndex.packageByName[pkg]
		if !ok {
			return errors.New("no such package: " + pkg)
		}

		generator, err := pd.getProjectGeneratorFunc(buildDir)
		if err != nil {
			return err
		}

		projectGenerators = append(projectGenerators, generator)
	}

	for _, generator := range projectGenerators {
		err = generator()
		if err != nil {
			return err
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
		if err := generatePackageSources(args); err != nil {
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
