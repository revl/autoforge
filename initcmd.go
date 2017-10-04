// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"log"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new workspace",
	Long: wrapText("The 'init' command prepares the current " +
		"(or the specified) directory for use by " + appName +
		" as a workspace directory."),
	Args: cobra.MaximumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := createWorkspace(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(initCmd)

	initCmd.Flags().SortFlags = false

	addWorkspaceDirFlag(initCmd)
	addPkgPathFlag(initCmd)
	addInstallDirFlag(initCmd)
}
