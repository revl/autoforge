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
	maketarget string, quiet bool) {

	fmt.Printf("Initializing a new workspace for %s...\n", appName)
	ip := initParams{OutputDir: "/home"}
	out, err := yaml.Marshal(&ip)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
}
