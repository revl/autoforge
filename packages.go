// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type packageDefinition struct {
	packageName string
	packageType string
	pathname    string
	requires    []string
	params      templateParams
}

func getRequiredField(pd *packageDefinition,
	fieldName string) (interface{}, error) {
	if value := pd.params[fieldName]; value != nil {
		return value, nil
	}
	return nil, errors.New(pd.pathname +
		": missing required field '" + fieldName + "'")
}

func getRequiredStringField(pd *packageDefinition,
	fieldName string) (string, error) {
	if value, err := getRequiredField(pd, fieldName); err != nil {
		return "", err
	} else if stringValue, ok := value.(string); ok {
		return stringValue, nil
	} else {
		return "", errors.New(pd.pathname +
			": '" + fieldName + "' field must be a string")
	}
}

func loadPackageDefinition(pathname string) (pd packageDefinition, err error) {
	data, err := ioutil.ReadFile(pathname)

	if err != nil {
		return
	}

	if err = yaml.Unmarshal(data, &pd.params); err != nil {
		errMessage := strings.TrimPrefix(err.Error(), "yaml: ")
		err = errors.New(pathname + ": " + errMessage)
		return
	}

	pd.packageName, err = getRequiredStringField(&pd, "name")
	if err != nil {
		return
	}

	pd.packageType, err = getRequiredStringField(&pd, "type")
	if err != nil {
		return
	}

	if requiredPackages := pd.params["requires"]; requiredPackages != nil {
		pkgList, ok := requiredPackages.([]interface{})
		if !ok {
			err = errors.New(pathname +
				": 'requires' must be a list")
			return
		}
		for _, pkgName := range pkgList {
			pd.requires = append(pd.requires, pkgName.(string))
		}
	}

	return
}

type packageIndex struct {
	packageByName   map[string]packageDefinition
	orderedPackages []packageDefinition
}

func getPackagePathFromEnvironment() (string, error) {
	if pkgpath := flags.pkgPath; pkgpath != "" {
		return pkgpath, nil
	}

	if pkgpath := os.Getenv(pkgPathEnvVar); pkgpath != "" {
		return pkgpath, nil
	}

	return "", errors.New("--pkgpath is not given and $" +
		pkgPathEnvVar + " is not defined")
}

func getPackagePathFromWorkspace() (string, error) {
	wp, err := readWorkspaceParams()
	if err != nil {
		return "", err
	}

	return wp.PkgPath, nil
}

func getPackagePathFromWorkspaceOrEnvironment() (string, error) {
	pkgpath, err := getPackagePathFromWorkspace()
	if pkgpath != "" && err == nil {
		return pkgpath, nil
	}

	return getPackagePathFromEnvironment()
}

func buildPackageIndex() (packageIndex, error) {
	var pi packageIndex

	pkgpath, err := getPackagePathFromWorkspaceOrEnvironment()

	if err != nil {
		return pi, err
	}

	pkgpathDirs := append(strings.Split(pkgpath, ":"),
		filepath.Join(filepath.Dir(os.Args[0]), "templates"))

	for _, pkgpathDir := range pkgpathDirs {
		dirEntries, _ := ioutil.ReadDir(pkgpathDir)

		for _, dirEntry := range dirEntries {
			dirEntryPathname := filepath.Join(pkgpathDir,
				dirEntry.Name(), dirEntry.Name()+".yaml")

			fileInfo, err := os.Stat(dirEntryPathname)
			if err != nil || !fileInfo.Mode().IsRegular() {
				continue
			}

			pd, err := loadPackageDefinition(dirEntryPathname)

			if err != nil {
				return packageIndex{}, err
			}

			pi.orderedPackages = append(pi.orderedPackages, pd)
		}
	}

	return pi, nil
}

func (index *packageIndex) printListOfPackages() {
	fmt.Println("List of packages:")

	for _, pd := range index.orderedPackages {
		fmt.Println(pd.packageName)
		fmt.Println(pd.packageType)
		for _, rp := range pd.requires {
			fmt.Println("-", rp)
		}
		fmt.Println()
	}
}
