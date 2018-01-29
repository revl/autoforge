// Copyright (C) 2017, 2018 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

func configurePackage(_ string, pd *packageDefinition) error {
	fmt.Println("[configure] " + pd.PackageName)
	return nil
}

func configurePackages(workspaceDir string, pkgNames []string) error {
	wp, err := readWorkspaceParams(workspaceDir)
	if err != nil {
		return err
	}

	pi, err := readPackageDefinitions(workspaceDir, wp)
	if err != nil {
		return err
	}

	if len(pkgNames) == 0 {
		privateDir := getPrivateDir(workspaceDir)

		selection, err := readPackageSelection(pi, privateDir)
		if err != nil {
			return err
		}

		for _, pd := range selection {
			configurePackage(workspaceDir, pd)
			if err != nil {
				return err
			}
		}
	} else {
		pd, err := pi.getPackageByName(pkgNames[0])
		if err != nil {
			return err
		}
		configurePackage(workspaceDir, pd)
	}

	return nil
}

// configureCmd represents the configure command
var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure all selected packages or the specified package",
	Args:  cobra.MaximumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		err := configurePackages(getWorkspaceDir(), args)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(configureCmd)

	configureCmd.Flags().SortFlags = false
	addQuietFlag(configureCmd)
	addWorkspaceDirFlag(configureCmd)
}
