// Copyright (C) 2017, 2018 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"io/ioutil"
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
