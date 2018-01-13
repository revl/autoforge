// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestNoPackages(t *testing.T) {
	pi, err := buildPackageIndex(packageDefinitionList{}, [][]string{})

	if err != nil {
		t.Error("Building index for an empty list returned an error")
	}

	if pi.packageByName == nil || pi.orderedPackages == nil ||
		len(pi.packageByName) != 0 || len(pi.orderedPackages) != 0 {
		t.Error("Index structures are not properly initialized")
	}
}

func dummyPackageDefinition(pkgName string) *packageDefinition {
	var pd packageDefinition
	pd.packageName = pkgName
	pd.pathname = filepath.Join(pkgName, packageDefinitionFilename)
	return &pd
}

func TestDuplicateDefinition(t *testing.T) {
	pkgList := packageDefinitionList{
		dummyPackageDefinition("base"),
		dummyPackageDefinition("client"),
		dummyPackageDefinition("base")}

	pi, err := buildPackageIndex(pkgList, [][]string{
		[]string{},
		[]string{},
		[]string{}})

	if pi != nil || err == nil || !strings.Contains(err.Error(),
		"duplicate package name: base") {
		t.Error("Package duplicate was not detected")
	}
}

func checkForCircularDependency(t *testing.T,
	pkgList packageDefinitionList,
	dependencies [][]string,
	cycle string) {
	_, err := buildPackageIndex(pkgList, dependencies)

	if err == nil {
		t.Error("Circular dependency was not detected")
	} else if !strings.Contains(err.Error(),
		"circular dependency detected: "+cycle) {
		t.Error("Unexpected circular dependency error: " +
			err.Error())
	}
}

func TestCircularDependency(t *testing.T) {
	checkForCircularDependency(t, packageDefinitionList{
		dummyPackageDefinition("a"),
		dummyPackageDefinition("b"),
		dummyPackageDefinition("c")},
		[][]string{
			[]string{"b"},
			[]string{"c"},
			[]string{"a"}},
		"a -> b -> c -> a")

	checkForCircularDependency(t, packageDefinitionList{
		dummyPackageDefinition("a"),
		dummyPackageDefinition("b"),
		dummyPackageDefinition("c"),
		dummyPackageDefinition("d")},
		[][]string{
			[]string{"b"},
			[]string{"c"},
			[]string{"b", "d"},
			[]string{}},
		"b -> c -> b")

	checkForCircularDependency(t, packageDefinitionList{
		dummyPackageDefinition("a"),
		dummyPackageDefinition("b")},
		[][]string{
			[]string{"b", "a"},
			[]string{}},
		"a -> a")
}

func TestDiamondDependency(t *testing.T) {
	pi, err := buildPackageIndex(packageDefinitionList{
		dummyPackageDefinition("d"),
		dummyPackageDefinition("b"),
		dummyPackageDefinition("c"),
		dummyPackageDefinition("a")},
		[][]string{
			[]string{"b", "c"},
			[]string{"a"},
			[]string{"a"},
			[]string{}},
	)

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
