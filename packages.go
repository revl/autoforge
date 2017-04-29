// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type packageDefinition struct {
	Name string `yaml:"name"`
}

func loadPackageDefinition(pathname string) packageDefinition {
	data, err := ioutil.ReadFile(pathname)

	if err != nil {
		panic(err)
	}

	var pd packageDefinition

	err = yaml.Unmarshal(data, &pd)

	if err != nil {
		panic(err)
	}

	return pd
}

type packageIndex struct {
	packageByName   map[string]packageDefinition
	orderedPackages []packageDefinition
}

func buildPackageIndex(pkgpath string) (packageIndex, error) {
	if len(pkgpath) == 0 {
		pkgpath = os.Getenv(pkgPathEnvVar)

		if len(pkgpath) == 0 {
			return packageIndex{}, errors.New(
				"-pkgpath is not given and $" +
					pkgPathEnvVar + " is not defined")
		}
	}

	for _, pkgpathDir := range strings.Split(pkgpath, ":") {
		dirEntries, _ := ioutil.ReadDir(pkgpathDir)

		for _, dirEntry := range dirEntries {
			dirEntryPathname := path.Join(pkgpathDir,
				dirEntry.Name(), dirEntry.Name()+".yaml")

			fileInfo, err := os.Stat(dirEntryPathname)
			if err != nil || !fileInfo.Mode().IsRegular() {
				continue
			}

			pd := loadPackageDefinition(dirEntryPathname)

			fmt.Println(pd.Name)
		}
	}

	return packageIndex{}, nil
}

func (pkgIndex *packageIndex) printListOfPackages() {
	fmt.Println("List of packages:")
}
