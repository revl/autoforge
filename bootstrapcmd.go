// Copyright (C) 2017, 2018 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"

	"github.com/spf13/cobra"
)

func bootstrapPackage(pd *packageDefinition) error {
	fmt.Println("[bootstrap] " + pd.PackageName)

	bootstrapCmd := exec.Command("./autogen.sh")
	bootstrapCmd.Dir = pd.packageDir()
	bootstrapCmd.Stdout = os.Stdout
	bootstrapCmd.Stderr = os.Stderr
	if err := bootstrapCmd.Run(); err != nil {
		return errors.New(path.Join(bootstrapCmd.Dir,
			"autogen.sh") + ": " + err.Error())
	}

	return nil
}

func bootstrapPackages(args []string) error {
	ws, err := loadWorkspace()
	if err != nil {
		return err
	}

	pi, err := readPackageDefinitions(ws.wp)
	if err != nil {
		return err
	}

	var selection packageDefinitionList

	if len(args) > 0 {
		selection, err = packageRangesToFlatSelection(pi, args)
	} else {
		selection, err = readPackageSelection(pi, ws.absPrivateDir)
	}
	if err != nil {
		return err
	}

	for _, pd := range selection {
		err = bootstrapPackage(pd)
		if err != nil {
			return err
		}
	}

	return nil
}

// bootstrapCmd represents the bootstrap command
var bootstrapCmd = &cobra.Command{
	Use: "bootstrap [package_range...]",
	Short: "Bootstrap all selected packages " +
		"or the specified package range",
	Run: func(_ *cobra.Command, args []string) {
		if err := bootstrapPackages(args); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(bootstrapCmd)

	bootstrapCmd.Flags().SortFlags = false
	addQuietFlag(bootstrapCmd)
	addWorkspaceDirFlag(bootstrapCmd)
}
