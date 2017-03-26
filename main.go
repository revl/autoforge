package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	// Handle panics by printing the error and exiting with return code 1.
	defer func() {
		if err := recover(); err != nil {
			fmt.Fprintf(os.Stderr, "autoforge: %s\n", err)

			os.Exit(1)
		}
	}()

	// Parse and process command line arguments.

	flag.Usage = func() {
		fmt.Printf("Usage: %s [options] [package_range]\n\n",
			os.Args[0])

		flag.PrintDefaults()
	}

	initFlag := flag.Bool("init", false, "initialize a new workspace")

	installdir := flag.String("installdir", "",
		"target directory for 'make install'")

	docdir := flag.String("docdir", "",
		"installation directory for documentation")

	maketarget := flag.String("maketarget", "help",
		"default makefile target")

	quiet := flag.Bool("quiet", false,
		"do not log progress to standard output")

	flag.Parse()

	if *initFlag {
		initializeWorkspace(*installdir, *docdir, *maketarget, *quiet)
	}
}
