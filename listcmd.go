// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"log"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Print the list of packages found in $" + pkgPathEnvVar,
	Args:  cobra.MaximumNArgs(0),
	Run: func(_ *cobra.Command, _ []string) {
		packageIndex, err := buildPackageIndex()
		if err != nil {
			log.Fatal(err)
		}

		packageIndex.printListOfPackages()
	},
}

func init() {
	RootCmd.AddCommand(listCmd)

	listCmd.Flags().SortFlags = false
	addWorkspaceDirFlag(listCmd)
	addPkgPathFlag(listCmd)
}
