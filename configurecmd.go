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

// configureEnv is a cache of environment variables that is
// reused when configuring multiple packages.
type configureEnv struct {
	envSansPkgConfigPath []string
	origPkgConfigPath    string // original value of PKG_CONFIG_PATH
}

const pkgConfigPathVarName = "PKG_CONFIG_PATH"

func dropEnvVar(env []string, i int) []string {
	if i < len(env)-1 {
		env[i] = env[len(env)-1]
	}
	return env[:len(env)-1]
}

func prepareConfigureEnv() *configureEnv {
	ce := &configureEnv{envSansPkgConfigPath: os.Environ()}

	for i, v := range ce.envSansPkgConfigPath {
		if strings.HasPrefix(v, pkgConfigPathVarName+"=") {
			ce.envSansPkgConfigPath = dropEnvVar(
				ce.envSansPkgConfigPath, i)
			ce.origPkgConfigPath = strings.SplitN(v, "=", 2)[1]
			break
		}
	}

	return ce
}

func (ce *configureEnv) makeEnv(buildDir string) ([]string, error) {
	absBuildDir, err := filepath.Abs(buildDir)
	if err != nil {
		return nil, err
	}

	configuredPackageDirs, err := ioutil.ReadDir(absBuildDir)
	if err != nil {
		return nil, err
	}

	var configuredPackagePathames []string

	for _, d := range configuredPackageDirs {
		configuredPackagePathames = append(
			configuredPackagePathames, absBuildDir+"/"+d.Name())
	}

	pkgConfigPath := strings.Join(configuredPackagePathames, ":")

	if len(pkgConfigPath) == 0 {
		if len(ce.origPkgConfigPath) == 0 {
			return ce.envSansPkgConfigPath, nil
		}
		pkgConfigPath = ce.origPkgConfigPath
	} else if len(ce.origPkgConfigPath) > 0 {
		pkgConfigPath += ":"
		pkgConfigPath += ce.origPkgConfigPath
	}

	return append(ce.envSansPkgConfigPath,
		pkgConfigPathVarName+"="+pkgConfigPath), nil
}

func configurePackage(privateDir string, wp *workspaceParams,
	pd *packageDefinition, cfgEnv *configureEnv) error {
	fmt.Println("[configure] " + pd.PackageName)

	conftab, err := readConftab(privateDir + "/" + conftabFilename)
	if err != nil {
		return err
	}

	pkgRootDir := getGeneratedPkgRootDir(privateDir)
	pkgDir := pkgRootDir + "/" + pd.PackageName

	buildDir := getBuildDir(privateDir, wp)

	env, err := cfgEnv.makeEnv(buildDir)
	if err != nil {
		return err
	}

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

func configurePackageSelection(wp *workspaceParams, privateDir string,
	selection packageDefinitionList) error {
	cfgEnv := prepareConfigureEnv()

	for _, pd := range selection {
		err := configurePackage(privateDir, wp, pd, cfgEnv)
		if err != nil {
			return err
		}
	}

	return nil
}

func configurePackages(workspaceDir string, args []string) error {
	wp, err := readWorkspaceParams(workspaceDir)
	if err != nil {
		return err
	}

	pi, err := readPackageDefinitions(workspaceDir, wp)
	if err != nil {
		return err
	}

	privateDir := getPrivateDir(workspaceDir)

	selection, err := readPackageSelection(pi, privateDir)
	if err != nil {
		return err
	}

	if len(args) > 0 {
		selectionFromArgs, err := packageRangesToFlatSelection(pi, args)
		if err != nil {
			return err
		}

		return configurePackageSelection(wp, privateDir,
			selectionFromArgs)
	}

	return configurePackageSelection(wp, privateDir, selection)
}

// configureCmd represents the configure command
var configureCmd = &cobra.Command{
	Use: "configure package_range...",
	Short: "Configure all selected packages " +
		"or the specified package range",
	Run: func(_ *cobra.Command, args []string) {
		err := configurePackages(getWorkspaceDir(), args)
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
