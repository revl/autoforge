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
	PkgPath    string `yaml:"pkgpath,omitempty"`
	InstallDir string `yaml:"installdir,omitempty"`
}

func createWorkspace() (*workspaceParams, error) {
	workspaceDir := "." + appName

	if flags.workspaceDir != "" {
		workspaceDir = filepath.Join(flags.workspaceDir, workspaceDir)
	}

	if _, err := os.Stat(workspaceDir); err == nil {
		return nil, errors.New("Workspace already initialized")
	}

	wp := workspaceParams{flags.pkgPath, flags.installDir}

	out, err := yaml.Marshal(&wp)
	if err != nil {
		return nil, err
	}

	err = os.Mkdir(workspaceDir, os.FileMode(0775))
	if err != nil {
		return nil, err
	}

	workspaceFile := filepath.Join(workspaceDir, "init.yaml")
	err = ioutil.WriteFile(workspaceFile, out, os.FileMode(0664))
	if err != nil {
		return nil, err
	}

	return &workspaceParams{}, nil
}
