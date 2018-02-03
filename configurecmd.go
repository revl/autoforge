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
	absBuildDir          string
	pkgAbsBuildDir       map[string]string
}

const pkgConfigPathVarName = "PKG_CONFIG_PATH"

func dropEnvVar(env []string, i int) []string {
	if i < len(env)-1 {
		env[i] = env[len(env)-1]
	}
	return env[:len(env)-1]
}

func prepareConfigureEnv(absBuildDir string) *configureEnv {
	ce := &configureEnv{
		os.Environ(),
		"",
		absBuildDir,
		map[string]string{}}

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

func (ce *configureEnv) addPackageBuildDir(pkgName string) {
	ce.pkgAbsBuildDir[pkgName] = ce.absBuildDir + "/" + pkgName
}

func (ce *configureEnv) makeEnv(pd *packageDefinition) []string {
	var configuredPackagePathames []string

	for _, dep := range pd.allRequired {
		depBuildDir, found := ce.pkgAbsBuildDir[dep.PackageName]
		if found {
			configuredPackagePathames = append(
				configuredPackagePathames, depBuildDir)
		}
	}

	pkgConfigPath := strings.Join(configuredPackagePathames, ":")

	if len(pkgConfigPath) == 0 {
		if len(ce.origPkgConfigPath) == 0 {
			return ce.envSansPkgConfigPath
		}
		pkgConfigPath = ce.origPkgConfigPath
	} else if len(ce.origPkgConfigPath) > 0 {
		pkgConfigPath += ":"
		pkgConfigPath += ce.origPkgConfigPath
	}

	return append(ce.envSansPkgConfigPath,
		pkgConfigPathVarName+"="+pkgConfigPath)
}

func configurePackage(privateDir string,
	pd *packageDefinition, cfgEnv *configureEnv,
	conftab *Conftab) error {
	fmt.Println("[configure] " + pd.PackageName)

	pkgRootDir := getGeneratedPkgRootDir(privateDir)
	pkgDir := pkgRootDir + "/" + pd.PackageName

	env := cfgEnv.makeEnv(pd)

	pkgBuildDir := cfgEnv.absBuildDir + "/" + pd.PackageName

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

	buildDir := getBuildDir(privateDir, wp)

	absBuildDir, err := filepath.Abs(buildDir)
	if err != nil {
		return err
	}

	configuredPackageDirs, err := ioutil.ReadDir(absBuildDir)
	if err != nil {
		return err
	}

	cfgEnv := prepareConfigureEnv(absBuildDir)

	// Register packages that already exist in the build directory.
	for _, dir := range configuredPackageDirs {
		cfgEnv.addPackageBuildDir(dir.Name())
	}

	var selection packageDefinitionList

	if len(args) > 0 {
		selection, err = packageRangesToFlatSelection(pi, args)
	} else {
		selection, err = readPackageSelection(pi, privateDir)
	}
	if err != nil {
		return err
	}

	for _, pd := range selection {
		cfgEnv.addPackageBuildDir(pd.PackageName)
	}

	conftab, err := readConftab(privateDir + "/" + conftabFilename)
	if err != nil {
		return err
	}

	for _, pd := range selection {
		err := configurePackage(privateDir, pd, cfgEnv, conftab)
		if err != nil {
			return err
		}
	}

	return nil
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