// Copyright (C) 2017, 2018 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"log"

	"github.com/spf13/cobra"
)

// menuCmd represents the menu command
var menuCmd = &cobra.Command{
	Use:   "menu",
	Short: "Print the list of packages found in $" + pkgPathEnvVar,
	Args:  cobra.MaximumNArgs(0),
	Run: func(_ *cobra.Command, _ []string) {
		workspaceDir, err := getWorkspaceDir()
		if err != nil {
			log.Fatal(err)
		}

		wp, err := readWorkspaceParams(workspaceDir)
		if err != nil {
			log.Fatal(err)
		}

		packageIndex, err := readPackageDefinitions(workspaceDir, wp)
		if err != nil {
			log.Fatal(err)
		}

		packageIndex.printListOfPackages()
	},
}

func init() {
	rootCmd.AddCommand(menuCmd)

	menuCmd.Flags().SortFlags = false
	addPkgPathFlag(menuCmd)
	addWorkspaceDirFlag(menuCmd)
}
