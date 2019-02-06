// Copyright (C) 2017, 2018 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

// +build !windows

package main

import (
	"bytes"
	"go/doc"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var appName = "autoforge"

var pkgPathEnvVar = "AUTOFORGE_PKG_PATH"

func wrapText(text string) string {
	var buffer bytes.Buffer

	doc.ToText(&buffer, text, "", "    ", 80)

	return buffer.String()
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   appName,
	Short: "Project generator for GNU Autotools",
}

func main() {
	// Suppress timestamps in the log messages.
	log.SetFlags(0)

	// Use application name for the log prefix.
	log.SetPrefix(appName + ": ")

	// Parse and process command line arguments.
	if rootCmd.Execute() != nil {
		os.Exit(1)
	}
}
