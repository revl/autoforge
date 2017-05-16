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
	Name     string   `yaml:"name"`
	Template string   `yaml:"template"`
	Requires []string `yaml:"requires"`
}

func loadPackageDefinition(pathname string) (pd packageDefinition, err error) {
	data, err := ioutil.ReadFile(pathname)

	if err != nil {
		return
	}

	err = yaml.Unmarshal(data, &pd)

	if err != nil {
		errMessage := strings.TrimPrefix(err.Error(), "yaml: ")
		err = errors.New(pathname + ": " + errMessage)
	}

	return
}

type packageIndex struct {
	packageByName   map[string]packageDefinition
	orderedPackages []packageDefinition
}

func buildPackageIndex(pkgpath string) (packageIndex, error) {
	var pi packageIndex

	if len(pkgpath) == 0 {
		pkgpath = os.Getenv(pkgPathEnvVar)

		if len(pkgpath) == 0 {
			return pi, errors.New("-pkgpath is not given and $" +
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

			pd, err := loadPackageDefinition(dirEntryPathname)

			if err != nil {
				panic(err)
			}

			pi.orderedPackages = append(pi.orderedPackages, pd)
		}
	}

	return pi, nil
}

func (index *packageIndex) printListOfPackages() {
	fmt.Println("List of packages:")

	for _, pd := range index.orderedPackages {
		fmt.Println(pd.Name)
		fmt.Println(pd.Template)
		for _, requiredPackage := range pd.Requires {
			fmt.Println("-", requiredPackage)
		}
	}
}
