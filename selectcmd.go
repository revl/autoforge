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

	var generators []func() error

	buildDir := filepath.Join(getWorkspaceDir(), "build")

	for _, pkg := range packages {
		pd, ok := packageIndex.packageByName[pkg]
		if !ok {
			return errors.New("no such package: " + pkg)
		}

		projectDir := filepath.Join(buildDir, pd.packageName)

		switch pd.packageType {
		case "app", "application":
			generators = append(generators, func() error {
				return generateBuildFilesFromEmbeddedTemplate(
					&appTemplate, projectDir, pd)
			})

		case "lib", "library":
			generators = append(generators, func() error {
				return generateBuildFilesFromEmbeddedTemplate(
					&libTemplate, projectDir, pd)
			})

		default:
			return errors.New(pkg + ": unknown package type '" +
				pd.packageType + "'")
		}
	}

	for _, generator := range generators {
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
