// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"errors"
	"log"

	"github.com/spf13/cobra"
)

func generatePackageSources(packages []string) error {
	packageIndex, err := buildPackageIndex()
	if err != nil {
		return err
	}

	for _, pkg := range packages {
		if _, ok := packageIndex.packageByName[pkg]; !ok {
			return errors.New("no such package: " + pkg)
		}
	}

	for _, pkg := range packages {
		pd := packageIndex.packageByName[pkg]

		err = generateBuildFilesFromEmbeddedTemplate(&appTemplate,
			"output-"+pd.packageName, pd)
		if err != nil {
			return err
		}
	}

	return nil
}

// pullCmd represents the pull command
var pullCmd = &cobra.Command{
	Use:   "pull [package_range...]",
	Short: "Generate Autotools files to build one or more packages",
	Run: func(_ *cobra.Command, args []string) {
		if err := generatePackageSources(args); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(pullCmd)

	pullCmd.Flags().SortFlags = false
	addQuietFlag(pullCmd)
	addWorkspaceDirFlag(pullCmd)
	addPkgPathFlag(pullCmd)
}
