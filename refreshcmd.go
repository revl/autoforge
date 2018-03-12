// Copyright (C) 2017, 2018 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"bufio"
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func readPackageSelection(pi *packageIndex, privateDir string) (
	packageDefinitionList, error) {
	file, err := os.Open(filepath.Join(privateDir,
		filenameForSelectedPackages))
	if err != nil {
		return nil, err
	}
	defer func() {
		closeErr := file.Close()
		if err == nil {
			err = closeErr
		}
	}()

	var selected packageDefinitionList

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		pkgName := scanner.Text()

		pd := pi.packageByName[pkgName]
		if pd == nil {
			return nil, errors.New("previously selected package '" +
				pkgName + "' could not be found")
		}

		selected = append(selected, pd)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return selected, nil
}

func refreshWorkspace() error {
	ws, err := loadWorkspace()
	if err != nil {
		return err
	}

	pi, err := readPackageDefinitions(ws.wp)
	if err != nil {
		return err
	}

	selection, err := readPackageSelection(pi, ws.absPrivateDir)
	if err != nil {
		return err
	}

	conftab, err := readConftab(filepath.Join(ws.absPrivateDir,
		conftabFilename))
	if err != nil {
		return err
	}

	return generateAndBootstrapPackages(ws, selection, conftab)
}

// refreshCmd represents the refresh command
var refreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Regenerate Autotools files in the current workspace",
	Args:  cobra.MaximumNArgs(0),
	Run: func(_ *cobra.Command, _ []string) {
		if err := refreshWorkspace(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(refreshCmd)

	refreshCmd.Flags().SortFlags = false
	addQuietFlag(refreshCmd)
	addWorkspaceDirFlag(refreshCmd)
	addNoBootstrapFlag(refreshCmd)
}
