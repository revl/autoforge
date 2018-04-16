// Copyright (C) 2017, 2018 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"errors"
	"io/ioutil"
	"log"
	"os"

	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

func initWorkspace() error {
	workspaceDir, err := getWorkspaceDir()
	if err != nil {
		return err
	}

	privateDir := getPrivateDir(workspaceDir)

	if _, err = os.Stat(privateDir); err == nil {
		return errors.New("workspace already initialized")
	}

	pkgpath, err := getPkgPathFlag()
	if err != nil {
		return err
	}
	if pkgpath == "" {
		pkgpath = os.Getenv(pkgPathEnvVar)
		if pkgpath == "" {
			return errors.New("--pkgpath is not given and $" +
				pkgPathEnvVar + " is not defined")
		}
	}

	buildDir, err := absIfNotEmpty(flags.buildDir)
	if err != nil {
		return err
	}

	installDir, err := absIfNotEmpty(flags.installDir)
	if err != nil {
		return err
	}

	wp := workspaceParams{flags.quiet, pkgpath,
		flags.makefile, flags.defaultMakeTarget,
		buildDir, installDir}

	out, err := yaml.Marshal(&wp)
	if err != nil {
		return err
	}

	err = os.Mkdir(privateDir, os.FileMode(0775))
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(getPathToSettings(privateDir),
		out, os.FileMode(0664))
	if err != nil {
		return err
	}

	return nil
}

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new workspace",
	Long: wrapText("The 'init' command prepares the current " +
		"(or the specified) directory for use by " + appName +
		" as a workspace directory."),
	Args: cobra.MaximumNArgs(0),
	Run: func(_ *cobra.Command, _ []string) {
		if err := initWorkspace(); err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().SortFlags = false

	addQuietFlag(initCmd)
	addPkgPathFlag(initCmd)
	addWorkspaceDirFlag(initCmd)
	addMakefileFlag(initCmd)
	addDefaultMakeTargetFlag(initCmd)
	addBuildDirFlag(initCmd)
	addInstallDirFlag(initCmd)
}
