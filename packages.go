// Copyright (C) 2017, 2018 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

var packageDefinitionFilename = appName + ".yaml"

type packageDefinition struct {
	PackageName  string
	description  string
	packageType  string
	pathname     string
	required     packageDefinitionList // Explicitly required packages
	allRequired  packageDefinitionList // Required + indirectly required
	uniqRequired packageDefinitionList // 'required' sans indirect reqs
	dependent    packageDefinitionList // Packages that depend on this one
	params       templateParams
}

type packageDefinitionList []*packageDefinition

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

func loadPackageDefinition(pathname string) (*packageDefinition, []string,
	error) {
	data, err := ioutil.ReadFile(pathname)
	if err != nil {
		return nil, nil, err
	}

	var params templateParams

	if err = yaml.Unmarshal(data, &params); err != nil {
		errMessage := strings.TrimPrefix(err.Error(), "yaml: ")
		err = errors.New(pathname + ": " + errMessage)
		return nil, nil, err
	}

	packageName, err := getRequiredStringField(pathname, params, "name")
	if err != nil {
		return nil, nil, err
	}

	description, err := getRequiredStringField(pathname, params,
		"description")
	if err != nil {
		return nil, nil, err
	}

	packageType, err := getRequiredStringField(pathname, params, "type")
	if err != nil {
		return nil, nil, err
	}

	_, err = getRequiredStringField(pathname, params, "version")
	if err != nil {
		return nil, nil, err
	}

	requires := []string{}

	if requiredPackages := params["requires"]; requiredPackages != nil {
		pkgList, ok := requiredPackages.([]interface{})
		if !ok {
			return nil, nil, errors.New(pathname +
				": 'requires' must be a list")
		}
		for _, pkgName := range pkgList {
			pkgNameStr, ok := pkgName.(string)
			if !ok {
				return nil, nil, errors.New(pathname +
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
		/*required*/ packageDefinitionList{},
		/*allRequired*/ packageDefinitionList{},
		/*uniqRequired*/ packageDefinitionList{},
		/*dependent*/ packageDefinitionList{},
		params}, requires, nil
}

type packageIndex struct {
	packageByName   map[string]*packageDefinition
	orderedPackages packageDefinitionList
}

func (pi *packageIndex) getPackageByName(pkgName string) (
	*packageDefinition, error) {
	if pd := pi.packageByName[pkgName]; pd != nil {
		return pd, nil
	}
	return nil, errors.New("no such package: " + pkgName)
}

func readPackageDefinitions(wp *workspaceParams) (*packageIndex, error) {
	var packages packageDefinitionList
	dependencies := [][]string{}

	pkgpath := flags.pkgPath
	if pkgpath == "" {
		pkgpath = wp.PkgPath
	} else {
		var err error
		pkgpath, err = getPkgPathFlag()
		if err != nil {
			return nil, err
		}
	}

	pkgpathDirs := append(strings.Split(pkgpath, ":"),
		path.Join(filepath.Dir(os.Args[0]), "templates"))

	for _, pkgpathDir := range pkgpathDirs {
		dirEntries, _ := ioutil.ReadDir(pkgpathDir)

		for _, dirEntry := range dirEntries {
			dirEntryPathname := path.Join(pkgpathDir,
				dirEntry.Name(), packageDefinitionFilename)

			fileInfo, err := os.Stat(dirEntryPathname)
			if err != nil || !fileInfo.Mode().IsRegular() {
				continue
			}

			pd, requires, err := loadPackageDefinition(
				dirEntryPathname)
			if err != nil {
				return nil, err
			}

			packages = append(packages, pd)
			dependencies = append(dependencies, requires)
		}
	}

	return buildPackageIndex(wp.Quiet, packages, dependencies)
}

type topologicalSorter struct {
	visited         map[*packageDefinition]int
	orderedPackages packageDefinitionList
}

const (
	unvisited = iota
	beingVisited
	visited
)

// cycle returns a string representing the cycle that
// has been detected in visit()
func (ts *topologicalSorter) cycle(pd, endp *packageDefinition) string {
	for _, dep := range pd.required {
		if ts.visited[dep] == beingVisited {
			if dep == endp {
				return pd.PackageName + " -> " +
					endp.PackageName
			}
			if cycle := ts.cycle(dep, endp); cycle != "" {
				return pd.PackageName + " -> " + cycle
			}
		}
	}
	return ""
}

func (ts *topologicalSorter) visit(pd *packageDefinition) error {
	switch ts.visited[pd] {
	case unvisited:
		ts.visited[pd] = beingVisited
		for _, dep := range pd.required {
			err := ts.visit(dep)
			if err != nil {
				return err
			}
		}
		ts.visited[pd] = visited
		ts.orderedPackages = append(ts.orderedPackages, pd)
	case beingVisited:
		return errors.New("circular dependency detected: " +
			ts.cycle(pd, pd))
	}
	return nil
}

// topologicalSort sorts the given package list using an algorithm based
// on depth-first search. Packages in the returned list are ordered so that
// all dependent packages come after the packages they depend on.
func topologicalSort(packages packageDefinitionList) (packageDefinitionList,
	error) {
	ts := topologicalSorter{make(map[*packageDefinition]int),
		packageDefinitionList{}}

	for _, pd := range packages {
		if ts.visited[pd] == unvisited {
			if err := ts.visit(pd); err != nil {
				return nil, err
			}
		}
	}

	return ts.orderedPackages, nil
}

// buildPackageIndex creates two types of structures for the
// input list of packages:
// 1. A map from package names to their definitions, and
// 2. A list of packages that contains a topological ordering
//    of the package dependency DAG.
func buildPackageIndex(quiet bool, packages packageDefinitionList,
	dependencies [][]string) (*packageIndex, error) {
	pi := &packageIndex{make(map[string]*packageDefinition),
		packageDefinitionList{}}

	// Create the packageByName index.
	for _, pd := range packages {
		// Having two different packages with the same name
		// is not allowed.
		if dup, ok := pi.packageByName[pd.PackageName]; ok {
			return nil, errors.New("duplicate package name: " +
				pd.PackageName + " (from " + pd.pathname +
				"); previously declared in " + dup.pathname)
		}
		pi.packageByName[pd.PackageName] = pd
	}

	// Resolve dependencies and compute the edges of the
	// reverse dependency DAG.
	for i, pd := range packages {
		for _, dep := range dependencies[i] {
			depp := pi.packageByName[dep]
			if depp == nil {
				return nil, errors.New("package " +
					pd.PackageName + " requires " +
					dep + ", which is not " +
					"available in the search path")
			}
			pd.required = append(pd.required, depp)
			depp.dependent = append(depp.dependent, pd)
		}
	}

	// Apply topological sorting to the dependency DAG so that
	// no package comes before the packages it depends on.
	var err error
	pi.orderedPackages, err = topologicalSort(packages)
	if err != nil {
		return nil, err
	}

	// For each package, find all of its dependencies,
	// including indirect ones. This computes the transitive
	// closure of the dependency DAG. Additionally, the second
	// nested loop also computes the transitive reduction
	// of the DAG.
	for _, pd := range pi.orderedPackages {
		allRequired := make(map[*packageDefinition]bool)

		// Recursion is not needed because the packages
		// are already ordered in such a way that the current
		// package never depends on those that follow it.
		for _, required := range pd.required {
			for _, dep := range required.allRequired {
				if !allRequired[dep] {
					pd.allRequired = append(
						pd.allRequired, dep)
					allRequired[dep] = true
				}
			}
		}

		// This loop cannot be merged with the previous one.
		// All indirect dependencies must be collected before
		// checking direct dependencies for redundancy.
		for _, required := range pd.required {
			if !allRequired[required] {
				pd.allRequired = append(
					pd.allRequired, required)
				allRequired[required] = true

				// Update the list of dependencies
				// exclusive to the current package.
				pd.uniqRequired = append(
					pd.uniqRequired, required)
			} else if !quiet {
				log.Printf("%s: redundant dependency on %s\n",
					pd.PackageName, required.PackageName)
			}
		}
	}

	return pi, nil
}

func establishDependenciesInSelection(selection packageDefinitionList,
	pi *packageIndex) map[*packageDefinition]packageDefinitionList {
	isSelected := make(map[*packageDefinition]bool)
	for _, pd := range selection {
		isSelected[pd] = true
	}

	// Build a graph of dependencies where the vertices are all
	// packages in the package index and the edges represent
	// dependencies of packages on *selected* packages.
	selectedPkgGraph := make(map[*packageDefinition]packageDefinitionList)

	for _, pd := range pi.orderedPackages {
		var selectedDeps packageDefinitionList

		added := make(map[*packageDefinition]bool)

		addDepIfNotAdded := func(dep *packageDefinition) {
			if !added[dep] {
				selectedDeps = append(selectedDeps, dep)
				added[dep] = true
			}
		}

		for _, dep := range pd.uniqRequired {
			if isSelected[dep] {
				addDepIfNotAdded(dep)
				continue
			}
			for _, indirectDep := range selectedPkgGraph[dep] {
				addDepIfNotAdded(indirectDep)
			}
		}

		selectedPkgGraph[pd] = selectedDeps
	}

	// Keep only selected vertices in the graph constructed above
	// and compute its transitive reduction.
	for _, pd := range pi.orderedPackages {
		if !isSelected[pd] {
			delete(selectedPkgGraph, pd)
			continue
		}

		indirectDeps := make(map[*packageDefinition]bool)

		for _, dep := range selectedPkgGraph[pd] {
			for _, indirectDep := range selectedPkgGraph[dep] {
				indirectDeps[indirectDep] = true
			}
		}

		var uniqueDeps packageDefinitionList
		for _, dep := range selectedPkgGraph[pd] {
			if !indirectDeps[dep] {
				uniqueDeps = append(uniqueDeps, dep)
			}
		}
		selectedPkgGraph[pd] = uniqueDeps
	}

	return selectedPkgGraph
}

func packageNames(pkgList packageDefinitionList) string {
	names := []string{}
	for _, pd := range pkgList {
		names = append(names, pd.PackageName)
	}
	return strings.Join(names, ", ")
}

func printListOfPackages(pkgList packageDefinitionList) {
	fmt.Println("List of packages:")

	for _, pd := range pkgList {
		fmt.Println("Name:", pd.PackageName)
		fmt.Println("Description:", pd.description)
		fmt.Println("Type:", pd.packageType)
		if len(pd.required) > 0 {
			fmt.Println("Requires:", packageNames(pd.required))
		}
		fmt.Println()
	}
}
