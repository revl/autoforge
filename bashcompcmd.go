// Copyright (C) 2017, 2018 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"os"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(&cobra.Command{
		Use:   "bashcomp",
		Short: "Print Bash completion script for " + appName,
		Long: wrapText("This command prints a Bash snippet, " +
			"which, when sourced or evaluated from .bashrc, " +
			"enables command completion for " + appName + "."),
		Args: cobra.MaximumNArgs(0),
		Run: func(_ *cobra.Command, _ []string) {
			RootCmd.GenBashCompletion(os.Stdout)
		},
	})
}
