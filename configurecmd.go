// Copyright (C) 2017, 2018 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func configurePackage(privateDir string, wp *workspaceParams,
	pd *packageDefinition, env []string) error {
	fmt.Println("[configure] " + pd.PackageName)

	conftab, err := readConftab(privateDir + "/" + conftabFilename)

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

	configureArgs := conftab.getConfigureArgs(pd.PackageName)
	configureArgs = append(configureArgs, "--quiet")

	configureCmd := exec.Command(configurePathname, configureArgs...)
	configureCmd.Dir = pkgBuildDir
	configureCmd.Stdout = os.Stdout
	configureCmd.Stderr = os.Stderr
	configureCmd.Env = env
	if err := configureCmd.Run(); err != nil {
		return errors.New(configurePathname + ": " + err.Error())
	}

	return nil
}

func updateEnviron(buildDir string) ([]string, error) {
	env := os.Environ()

	absBuildDir, err := filepath.Abs(buildDir)
	if err != nil {
		return nil, err
	}

	configuredPackageDirs, err := ioutil.ReadDir(absBuildDir)
	if err != nil {
		return nil, err
	}

	if len(configuredPackageDirs) == 0 {
		return env, nil
	}

	var configuredPackagePathames []string

	for _, d := range configuredPackageDirs {
		configuredPackagePathames = append(
			configuredPackagePathames, absBuildDir+"/"+d.Name())
	}

	uninstalledConfigPath := strings.Join(configuredPackagePathames, ":")

	const pkgConfigPathVarName = "PKG_CONFIG_PATH"

	for i, v := range env {
		if strings.HasPrefix(v, pkgConfigPathVarName+"=") {
			split := strings.SplitAfterN(v, "=", 2)
			newValue := split[0] + uninstalledConfigPath
			if len(split[1]) > 0 {
				newValue += ":"
				newValue += split[1]
			}
			env[i] = newValue
			return env, nil
		}
	}

	return append(env, pkgConfigPathVarName+"="+uninstalledConfigPath), nil
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

	privateDir := getPrivateDir(workspaceDir)

	buildDir := getBuildDir(privateDir, wp)

	env, err := updateEnviron(buildDir)
	if err != nil {
		return err
	}

	if len(pkgNames) > 0 {
		pd, err := pi.getPackageByName(pkgNames[0])
		if err != nil {
			return err
		}
		return configurePackage(privateDir, wp, pd, env)
	}

	selection, err := readPackageSelection(pi, privateDir)
	if err != nil {
		return err
	}

	for _, pd := range selection {
		err = configurePackage(privateDir, wp, pd, env)
		if err != nil {
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
