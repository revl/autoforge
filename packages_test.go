// Copyright (C) 2017, 2018 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestNoPackages(t *testing.T) {
	pi, err := buildPackageIndex(false,
		packageDefinitionList{}, [][]string{})

	if err != nil {
		t.Error("Building index for an empty list returned an error")
	}

	if pi.packageByName == nil || pi.orderedPackages == nil ||
		len(pi.packageByName) != 0 || len(pi.orderedPackages) != 0 {
		t.Error("Index structures are not properly initialized")
	}
}

func makePackageIndexForTesting(packagesAndDependencies []string, quiet bool) (
	*packageIndex, error) {
	var packages packageDefinitionList
	var deps [][]string

	for _, packageLine := range packagesAndDependencies {
		split := strings.SplitN(packageLine, ":", 2)

		packages = append(packages, &packageDefinition{
			PackageName: split[0],
			pathname: filepath.Join(split[0],
				packageDefinitionFilename)})

		if len(split) > 1 {
			deps = append(deps, strings.Split(split[1], ","))
		} else {
			deps = append(deps, []string{})
		}
	}

	return buildPackageIndex(quiet, packages, deps)
}

func TestDuplicateDefinition(t *testing.T) {
	pi, err := makePackageIndexForTesting([]string{
		"base", "client", "base"}, false)

	if pi != nil || err == nil || !strings.Contains(err.Error(),
		"duplicate package name: base") {
		t.Error("Package duplicate was not detected")
	}
}

func confirmCircularDependencyError(t *testing.T, err error, cycle string) {
	if err == nil {
		t.Error("Circular dependency was not detected")
	} else if !strings.Contains(err.Error(),
		"circular dependency detected: "+cycle) {
		t.Error("Unexpected circular dependency error: " +
			err.Error())
	}
}

func TestCircularDependency(t *testing.T) {
	_, err := makePackageIndexForTesting([]string{
		"a:b", "b:c", "c:a"}, false)

	confirmCircularDependencyError(t, err, "a -> b -> c -> a")

	_, err = makePackageIndexForTesting([]string{
		"a:b", "b:c", "c:b,d", "d"}, false)

	confirmCircularDependencyError(t, err, "b -> c -> b")

	_, err = makePackageIndexForTesting([]string{
		"a:b,a", "b"}, false)

	confirmCircularDependencyError(t, err, "a -> a")
}

func TestDiamondDependency(t *testing.T) {
	pi, err := makePackageIndexForTesting([]string{
		"d:b,c", "b:a", "c:a", "a"}, false)

	if err != nil {
		t.Error("Unexpected error")
	}

	if len(pi.packageByName) != len(pi.orderedPackages) {
		t.Error("Index size mismatch")
	}

	packageOrder := packageNames(pi.orderedPackages)
	if packageOrder != "a, b, c, d" {
		t.Error("Invalid package order: " + packageOrder)
	}

	checkIndirectDependencies := func(pkgName, expectedDeps string) {
		deps := packageNames(pi.packageByName[pkgName].allRequired)
		if deps != expectedDeps {
			t.Error("Indirect dependencies for " + pkgName +
				" do not match: expected=" + expectedDeps +
				"; actual=" + deps)
		}
	}

	checkIndirectDependencies("a", "")
	checkIndirectDependencies("b", "a")
	checkIndirectDependencies("c", "a")
	checkIndirectDependencies("d", "a, b, c")
}

func checkSelectionGraph(t *testing.T,
	packagesAndDependencies []string,
	expected map[string]string) {
	pi, err := makePackageIndexForTesting(packagesAndDependencies, true)

	if err != nil {
		t.Error("Unexpected error")
	}

	var selection packageDefinitionList
	for name := range expected {
		pd, found := pi.packageByName[name]
		if !found {
			t.Error("Invalid test case inputs")
		}
		selection = append(selection, pd)
	}

	selectionGraph := establishDependenciesInSelection(selection, pi)

	if len(selectionGraph) != len(expected) {
		t.Error("Unexpected number of selected vertices")
	}

	for pd, deps := range selectionGraph {
		name := pd.PackageName
		expectedDepNames, match := expected[name]
		if !match {
			t.Error("Unexpected package returned: " + name)
		}

		var depNames string
		for _, dep := range deps {
			if depNames != "" {
				depNames += ", "
			}
			depNames += dep.PackageName
		}

		if depNames != expectedDepNames {
			t.Error("Unexpected dependencies for " + name + ": " +
				depNames + "; expected: " + expectedDepNames)
		}
	}
}

func TestSelectionGraph(t *testing.T) {
	checkSelectionGraph(t,
		[]string{"d:a,b,c", "b:a", "c:a", "a"},
		map[string]string{"a": "", "d": "a"})

	packageDependencies := []string{
		"a",
		"b:a",
		"c:b",
		"d:b",
		"e:c",
		"f:b",
		"g:e,f,d",
		"h:g",
		"i:f,h",
		"j:i,h",
	}

	checkSelectionGraph(t,
		packageDependencies,
		map[string]string{
			"b": "",
			"c": "b",
			"d": "b",
			"g": "c, d",
			"i": "g",
		})

	checkSelectionGraph(t,
		packageDependencies,
		map[string]string{
			"a": "",
			"b": "a",
			"c": "b",
			"d": "b",
			"e": "c",
			"f": "b",
			"g": "e, f, d",
			"h": "g",
			"i": "h",
			"j": "i",
		})
}
