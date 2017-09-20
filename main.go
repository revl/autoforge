// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

var appName = "autoforge"

var pkgPathEnvVar = "AUTOFORGE_PKG_PATH"

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   appName,
	Short: "Project generator for GNU Autotools",
}

func main() {
	// Suppress timestamps in the log messages.
	log.SetFlags(0)

	// Use application name for the log prefix.
	log.SetPrefix(appName + ": ")

	// Handle panics by printing the error and exiting with return code 1.
	defer func() {
		if r := recover(); r != nil {
			log.Fatal(r)
		}
	}()

	// Parse and process command line arguments.
	if RootCmd.Execute() != nil {
		os.Exit(1)
	}
}
