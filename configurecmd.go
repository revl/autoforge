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
	"path/filepath"

	"github.com/spf13/cobra"
)

func configurePackage(workspaceDir string, wp *workspaceParams,
	pd *packageDefinition) error {
	fmt.Println("[configure] " + pd.PackageName)

	privateDir := getPrivateDir(workspaceDir)

	pkgRootDir := getGeneratedPkgRootDir(privateDir)
	pkgDir := pkgRootDir + "/" + pd.PackageName

	buildDir := getBuildDir(privateDir, wp)
	pkgBuildDir := buildDir + "/" + pd.PackageName

	relPkgSrcDir, err := filepath.Rel(pkgBuildDir, pkgDir)
	if err != nil {
		relPkgSrcDir, err = filepath.Abs(pkgDir)
		if err != nil {
			return err
		}
	}

	err = os.MkdirAll(pkgBuildDir, os.FileMode(0775))
	if err != nil {
		return nil
	}

	configurePathname := relPkgSrcDir + "/configure"

	configureCmd := exec.Command(configurePathname, "--quiet")
	configureCmd.Dir = pkgBuildDir
	configureCmd.Stdout = os.Stdout
	configureCmd.Stderr = os.Stderr
	if err := configureCmd.Run(); err != nil {
		return errors.New(configurePathname + ": " + err.Error())
	}

	return nil
}

func configurePackages(pkgNames []string) error {
	workspaceDir := getWorkspaceDir()

	wp, err := readWorkspaceParams(workspaceDir)
	if err != nil {
		return err
	}

	pi, err := readPackageDefinitions(workspaceDir, wp)
	if err != nil {
		return err
	}

	if len(pkgNames) > 0 {
		pd, err := pi.getPackageByName(pkgNames[0])
		if err != nil {
			return err
		}
		return configurePackage(workspaceDir, wp, pd)
	}

	privateDir := getPrivateDir(workspaceDir)

	selection, err := readPackageSelection(pi, privateDir)
	if err != nil {
		return err
	}

	for _, pd := range selection {
		if err = configurePackage(workspaceDir, wp, pd); err != nil {
			return err
		}
	}

	return nil
}

// configureCmd represents the configure command
var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure all selected packages or the specified package",
	Args:  cobra.MaximumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		err := configurePackages(args)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(configureCmd)

	configureCmd.Flags().SortFlags = false
	addQuietFlag(configureCmd)
	addWorkspaceDirFlag(configureCmd)
}
