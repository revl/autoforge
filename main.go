// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"os"
)

var appName = "autoforge"

var pkgPathEnvVar = "AUTOFORGE_PKG_PATH"

func printErrorAndExit(err string) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", appName, err)

	os.Exit(1)
}

func main() {
	// Handle panics by printing the error and exiting with return code 1.
	defer func() {
		if err := recover(); err != nil {
			printErrorAndExit(err.(string))
		}
	}()

	// Parse and process command line arguments.

	flag.Usage = func() {
		fmt.Printf("Usage: %s [options] [package_range]\n\n", appName)

		flag.PrintDefaults()
	}

	initFlag := flag.Bool("init", false, "initialize a new workspace")

	query := flag.Bool("query", false,
		"print the list of packages found in $"+pkgPathEnvVar)

	installdir := flag.String("installdir", "",
		"target directory for 'make install'")

	docdir := flag.String("docdir", "",
		"installation directory for documentation")

	maketarget := flag.String("maketarget", "help",
		"default makefile target")

	quiet := flag.Bool("quiet", false,
		"do not display progress and result of operation")

	pkgpath := flag.String("pkgpath", "",
		"the list of directories where to search for packages")

	workspacedir := flag.String("workspacedir", ".",
		"pathname of the workspace directory")

	flag.Parse()

	var err error

	switch {
	case *initFlag:
		initializeWorkspace(*workspacedir, *pkgpath, *installdir,
			*docdir, *maketarget, *quiet)
	case *query:
		err = queryPackages(*workspacedir, *pkgpath)
	}

	if err != nil {
		printErrorAndExit(err.Error())
	}
}
