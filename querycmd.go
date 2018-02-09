// Copyright (C) 2017, 2018 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"log"

	"github.com/spf13/cobra"
)

func queryPackages(args []string) error {
	workspaceDir, err := getWorkspaceDir()
	if err != nil {
		log.Fatal(err)
	}

	wp, err := readWorkspaceParams(workspaceDir)
	if err != nil {
		log.Fatal(err)
	}

	pi, err := readPackageDefinitions(workspaceDir, wp)
	if err != nil {
		log.Fatal(err)
	}

	if len(args) > 0 {
		selection, err := packageRangesToFlatSelection(pi, args)
		if err != nil {
			return err
		}
		printListOfPackages(selection)
	} else {
		printListOfPackages(pi.orderedPackages)
	}

	return nil
}

// queryCmd represents the query command
var queryCmd = &cobra.Command{
	Use:   "query [package_range...]",
	Short: "Print the list of packages found in $" + pkgPathEnvVar,
	Run: func(_ *cobra.Command, args []string) {
		if err := queryPackages(args); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(queryCmd)

	queryCmd.Flags().SortFlags = false
	addPkgPathFlag(queryCmd)
	addWorkspaceDirFlag(queryCmd)
}
