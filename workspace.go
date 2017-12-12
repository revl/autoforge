// Copyright (C) 2017 Damon Revoe. All rights reserved.
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
	PkgPath    string `yaml:"pkgpath"`
	InstallDir string `yaml:"installdir,omitempty"`
}

func getWorkspaceDir() string {
	workspaceDir := "." + appName

	if flags.workspaceDir != "" {
		workspaceDir = filepath.Join(flags.workspaceDir, workspaceDir)
	}

	return workspaceDir
}

func getPathToSettings(workspaceDir string) string {
	return filepath.Join(workspaceDir, "settings.yaml")
}

func createWorkspace() (*workspaceParams, error) {
	workspaceDir := getWorkspaceDir()

	if _, err := os.Stat(workspaceDir); err == nil {
		return nil, errors.New("workspace already initialized")
	}

	pkgpath, err := getPackagePathFromEnvironment()

	if err != nil {
		return nil, err
	}

	wp := workspaceParams{pkgpath, flags.installDir}

	out, err := yaml.Marshal(&wp)
	if err != nil {
		return nil, err
	}

	err = os.Mkdir(workspaceDir, os.FileMode(0775))
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(getPathToSettings(workspaceDir),
		out, os.FileMode(0664))
	if err != nil {
		return nil, err
	}

	return &workspaceParams{}, nil
}

func readWorkspaceParams() (*workspaceParams, error) {
	in, err := ioutil.ReadFile(getPathToSettings(getWorkspaceDir()))
	if err != nil {
		return nil, err
	}

	var wp workspaceParams
	err = yaml.Unmarshal(in, &wp)

	return &wp, err
}
