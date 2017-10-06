// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"log"

	"github.com/spf13/cobra"
)

func listPackages() error {
	packageIndex, err := buildPackageIndex()

	if err != nil {
		return err
	}

	packageIndex.printListOfPackages()

	return nil
}

// listCmd represents the init command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Print the list of packages found in $" + pkgPathEnvVar,
	Args:  cobra.MaximumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if err := listPackages(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(listCmd)

	listCmd.Flags().SortFlags = false
	addWorkspaceDirFlag(listCmd)
	addPkgPathFlag(listCmd)
}
