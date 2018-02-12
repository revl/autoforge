// Copyright (C) 2017, 2018 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type workspaceParams struct {
	Quiet             bool   `yaml:"quiet"`
	PkgPath           string `yaml:"pkgpath"`
	Makefile          string `yaml:"makefile,omitempty"`
	DefaultMakeTarget string `yaml:"default-target,omitempty"`
	BuildDir          string `yaml:"builddir,omitempty"`
	InstallDir        string `yaml:"installdir,omitempty"`
}

func getWorkspaceDir() (string, error) {
	return filepath.Abs(flags.workspaceDir)
}

var privateDirName = "." + appName

func getPrivateDir(workspaceDir string) string {
	return filepath.Join(workspaceDir, privateDirName)
}

func getPathToSettings(privateDir string) string {
	return filepath.Join(privateDir, "settings.yaml")
}

func createWorkspace() (*workspaceParams, error) {
	workspaceDir, err := getWorkspaceDir()
	if err != nil {
		return nil, err
	}

	privateDir := getPrivateDir(workspaceDir)

	if _, err = os.Stat(privateDir); err == nil {
		return nil, errors.New("workspace already initialized")
	}

	pkgpath, err := getPackagePathFromEnvironment()
	if err != nil {
		return nil, err
	}

	buildDir, err := absIfNotEmpty(flags.buildDir)
	if err != nil {
		return nil, err
	}

	installDir, err := absIfNotEmpty(flags.installDir)
	if err != nil {
		return nil, err
	}

	wp := workspaceParams{flags.quiet, pkgpath,
		flags.makefile, flags.defaultMakeTarget,
		buildDir, installDir}

	out, err := yaml.Marshal(&wp)
	if err != nil {
		return nil, err
	}

	err = os.Mkdir(privateDir, os.FileMode(0775))
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(getPathToSettings(privateDir),
		out, os.FileMode(0664))
	if err != nil {
		return nil, err
	}

	return &workspaceParams{}, nil
}

func readWorkspaceParams(workspaceDir string) (*workspaceParams, error) {
	in, err := ioutil.ReadFile(
		getPathToSettings(getPrivateDir(workspaceDir)))
	if err != nil {
		return nil, err
	}

	var wp workspaceParams
	err = yaml.Unmarshal(in, &wp)

	return &wp, err
}
