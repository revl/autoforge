// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
)

type initParams struct {
	OutputDir string `yaml:"outputdir,omitempty"`
}

func loadInitParams() initParams {
	return initParams{}
}

func initializeWorkspace(workspacedir, pkgpath, installdir, docdir,
	maketarget string, quiet bool) error {

	fmt.Printf("Initializing a new workspace for %s...\n", appName)

	ip := initParams{OutputDir: "/home"}

	out, err := yaml.Marshal(&ip)
	if err != nil {
		return err
	}

	fmt.Println(string(out))

	return nil
}
