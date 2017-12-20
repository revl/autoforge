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

var packageDefinitionFilename = appName + ".yaml"

type packageDefinition struct {
	packageName string
	description string
	packageType string
	pathname    string
	requires    []string
	uniqueReq   map[string]bool
	allReq      []string
	requiredBy  []string
	params      templateParams
}

func getRequiredField(pathname string, params templateParams,
	fieldName string) (interface{}, error) {
	if value := params[fieldName]; value != nil {
		return value, nil
	}
	return nil, errors.New(pathname +
		": missing required field '" + fieldName + "'")
}

func getRequiredStringField(pathname string, params templateParams,
	fieldName string) (string, error) {
	if value, err := getRequiredField(pathname,
		params, fieldName); err != nil {
		return "", err
	} else if stringValue, ok := value.(string); ok {
		return stringValue, nil
	} else {
		return "", errors.New(pathname +
			": '" + fieldName + "' field must be a string")
	}
}

func loadPackageDefinition(pathname string) (*packageDefinition, error) {
	data, err := ioutil.ReadFile(pathname)
	if err != nil {
		return nil, err
	}

	var params templateParams

	if err = yaml.Unmarshal(data, &params); err != nil {
		errMessage := strings.TrimPrefix(err.Error(), "yaml: ")
		err = errors.New(pathname + ": " + errMessage)
		return nil, err
	}

	packageName, err := getRequiredStringField(pathname, params, "name")
	if err != nil {
		return nil, err
	}

	description, err := getRequiredStringField(pathname, params,
		"description")
	if err != nil {
		return nil, err
	}

	packageType, err := getRequiredStringField(pathname, params, "type")
	if err != nil {
		return nil, err
	}

	_, err = getRequiredStringField(pathname, params, "version")
	if err != nil {
		return nil, err
	}

	requires := []string{}

	if requiredPackages := params["requires"]; requiredPackages != nil {
		pkgList, ok := requiredPackages.([]interface{})
		if !ok {
			return nil, errors.New(pathname +
				": 'requires' must be a list")
		}
		for _, pkgName := range pkgList {
			pkgNameStr, ok := pkgName.(string)
			if !ok {
				return nil, errors.New(pathname +
					": 'requires' must be " +
					"a list of strings")
			}
			requires = append(requires, pkgNameStr)
		}
	}

	return &packageDefinition{
		packageName,
		description,
		packageType,
		pathname,
		requires,
		/*uniqueReq*/ make(map[string]bool),
		/*allReq*/ []string{},
		/*requiredBy*/ []string{},
		params}, nil
}

type packageDefinitionList []*packageDefinition

type packageIndex struct {
	packageByName   map[string]*packageDefinition
	orderedPackages packageDefinitionList
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

func readPackageDefinitions() (*packageIndex, error) {
	var packages packageDefinitionList

	pkgpath, err := getPackagePathFromWorkspaceOrEnvironment()
	if err != nil {
		return nil, err
	}

	pkgpathDirs := append(strings.Split(pkgpath, ":"),
		filepath.Join(filepath.Dir(os.Args[0]), "templates"))

	for _, pkgpathDir := range pkgpathDirs {
		dirEntries, _ := ioutil.ReadDir(pkgpathDir)

		for _, dirEntry := range dirEntries {
			dirEntryPathname := filepath.Join(pkgpathDir,
				dirEntry.Name(), packageDefinitionFilename)

			fileInfo, err := os.Stat(dirEntryPathname)
			if err != nil || !fileInfo.Mode().IsRegular() {
				continue
			}

			pd, err := loadPackageDefinition(dirEntryPathname)
			if err != nil {
				return nil, err
			}

			packages = append(packages, pd)
		}
	}

	return buildPackageIndex(packages)
}

func buildPackageIndex(packages packageDefinitionList) (*packageIndex, error) {
	pi := &packageIndex{make(map[string]*packageDefinition),
		packageDefinitionList{}}

	for _, pd := range packages {
		existingPackage, ok := pi.packageByName[pd.packageName]
		if ok {
			return nil, errors.New("duplicate " +
				"package name: " + pd.packageName +
				" (from " + pd.pathname +
				"); previously declared in " +
				existingPackage.pathname)
		}
		pi.packageByName[pd.packageName] = pd
	}

	// Queue for ordering packages from least dependent to most dependent.
	queue := []string{}

	// Resolve package dependencies.
	for pkgName, pd := range pi.packageByName {
		if len(pd.requires) == 0 {
			queue = append(queue, pkgName)
			continue
		}

		for _, dependency := range pd.requires {
			requiredPackage := pi.packageByName[dependency]
			if requiredPackage == nil {
				return nil, errors.New("package " +
					pkgName + " requires " +
					dependency + ", which is not " +
					"available in the search path")
			}
			requiredPackage.requiredBy =
				append(requiredPackage.requiredBy, pkgName)

			pd.uniqueReq[dependency] = true
		}
	}

	for len(queue) > 0 {
		pkgName := queue[0]
		queue = queue[1:]

		pd := pi.packageByName[pkgName]

		pi.orderedPackages = append(pi.orderedPackages, pd)

		for _, requiredBy := range pd.requiredBy {
			dependentPackage := pi.packageByName[requiredBy]

			delete(dependentPackage.uniqueReq, pkgName)

			if len(dependentPackage.uniqueReq) == 0 {
				queue = append(queue,
					dependentPackage.packageName)
			}
		}
	}

	// For each package, find all of its indirect dependencies.
	for _, pd := range pi.orderedPackages {
		added := make(map[string]bool)

		addDep := func(pkgName string) {
			if !added[pkgName] {
				pd.allReq = append(pd.allReq, pkgName)
				added[pkgName] = true
			}
		}

		// Recursion is not needed because the packages
		// are already ordered in such a way that the current
		// package never depends on those that follow it.
		for _, required := range pd.requires {
			for _, dep := range pi.packageByName[required].allReq {
				addDep(dep)
			}
			addDep(required)
		}
	}

	return pi, nil
}

func (index *packageIndex) printListOfPackages() {
	fmt.Println("List of packages:")

	for _, pd := range index.orderedPackages {
		fmt.Println("Name:", pd.packageName)
		fmt.Println("Description:", pd.description)
		fmt.Println("Type:", pd.packageType)
		fmt.Println("Requires:", strings.Join(pd.requires, ","))
		fmt.Println()
	}
}
