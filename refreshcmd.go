// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"log"

	"github.com/spf13/cobra"
)

// RefreshCmd represents the refresh command
var refreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Regenerate Autotools files in the current workspace",
	Run: func(_ *cobra.Command, args []string) {
		if err := generateAndBootstrapPackages(getWorkspaceDir(),
			[]string{}); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(refreshCmd)

	refreshCmd.Flags().SortFlags = false
	addQuietFlag(refreshCmd)
	addWorkspaceDirFlag(refreshCmd)
}
