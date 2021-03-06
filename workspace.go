// Copyright (C) 2017, 2018 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"io/ioutil"
	"path"
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

type workspace struct {
	absDir        string
	absPrivateDir string
	wp            *workspaceParams
}

func getWorkspaceDir() (string, error) {
	return filepath.Abs(flags.workspaceDir)
}

var privateDirName = "." + appName

func getPrivateDir(workspaceDir string) string {
	return path.Join(workspaceDir, privateDirName)
}

func getPathToSettings(privateDir string) string {
	return path.Join(privateDir, "settings.yaml")
}

func loadWorkspace() (*workspace, error) {
	workspaceDir, err := getWorkspaceDir()
	if err != nil {
		return nil, err
	}

	privateDir := getPrivateDir(workspaceDir)

	in, err := ioutil.ReadFile(getPathToSettings(privateDir))
	if err != nil {
		return nil, err
	}

	var wp workspaceParams
	err = yaml.Unmarshal(in, &wp)

	return &workspace{workspaceDir, privateDir, &wp}, err
}

var pkgDirName = "packages"

// generatedPkgRootDir returns the absolute pathname of the
// directory where source files for Autotools are generated.
func (ws *workspace) generatedPkgRootDir() string {
	return path.Join(ws.absPrivateDir, pkgDirName)
}

// buildDir returns the absolute pathname of the directory
// where the packages are configured and built.
func (ws *workspace) buildDir() string {
	if ws.wp.BuildDir != "" {
		return ws.wp.BuildDir
	}
	return path.Join(ws.absPrivateDir, "build")
}

// installDir returns the absolute pathname of the directory
// where to install the built binaries and library headers.
func (ws *workspace) installDir() string {
	if flags.installDir != "" {
		return flags.installDir
	}
	return ws.absDir
}

// relativeToWorkspace returns an equivalent of 'absPath'
// that is relative to the workspace directory.
func (ws *workspace) relativeToWorkspace(absPath string) string {
	return relativeIfShorter(ws.absDir, absPath)
}

// pkgRootDirRelativeToWorkspace returns the pathname of the
// directory with generated package sources relative to the
// workspace directory.
func (ws *workspace) pkgRootDirRelativeToWorkspace() string {
	return ws.relativeToWorkspace(ws.generatedPkgRootDir())
}

// buildDirRelativeToWorkspace returns the pathname of the
// build directory relative to the workspace directory.
func (ws *workspace) buildDirRelativeToWorkspace() string {
	return ws.relativeToWorkspace(ws.buildDir())
}
