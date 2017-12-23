// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestNoPackages(t *testing.T) {
	pi, err := buildPackageIndex(packageDefinitionList{})

	if err != nil {
		t.Error("Building index for an empty list returned an error")
	}

	if pi.packageByName == nil || pi.orderedPackages == nil ||
		len(pi.packageByName) != 0 || len(pi.orderedPackages) != 0 {
		t.Error("Index structures are not properly initialized")
	}
}

func dummyPackageDefinition(pkgName string, dep ...string) *packageDefinition {
	var pd packageDefinition
	pd.packageName = pkgName
	pd.pathname = filepath.Join(pkgName, packageDefinitionFilename)
	pd.requires = dep
	return &pd
}

func TestDuplicateDefinition(t *testing.T) {
	pkgList := packageDefinitionList{
		dummyPackageDefinition("base"),
		dummyPackageDefinition("client"),
		dummyPackageDefinition("base"),
	}

	pi, err := buildPackageIndex(pkgList)

	if pi != nil || err == nil || !strings.Contains(err.Error(),
		"duplicate package name: base") {
		t.Error("Package duplicate was not detected")
	}
}

func checkForCircularDependency(t *testing.T,
	pkgList packageDefinitionList, cycle string) {
	_, err := buildPackageIndex(pkgList)

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
		dummyPackageDefinition("a", "b"),
		dummyPackageDefinition("b", "c"),
		dummyPackageDefinition("c", "a"),
	}, "a -> b -> c -> a")

	checkForCircularDependency(t, packageDefinitionList{
		dummyPackageDefinition("a", "b"),
		dummyPackageDefinition("b", "c"),
		dummyPackageDefinition("c", "b", "d"),
		dummyPackageDefinition("d"),
	}, "b -> c -> b")

	checkForCircularDependency(t, packageDefinitionList{
		dummyPackageDefinition("a", "b", "a"),
		dummyPackageDefinition("b"),
	}, "a -> a")
}

func packageNames(pkgList packageDefinitionList) string {
	names := []string{}
	for _, pd := range pkgList {
		names = append(names, pd.packageName)
	}
	return fmt.Sprintf("%v", names)
}

func TestDiamondDependency(t *testing.T) {
	pi, err := buildPackageIndex(packageDefinitionList{
		dummyPackageDefinition("d", "b", "c"),
		dummyPackageDefinition("b", "a"),
		dummyPackageDefinition("c", "a"),
		dummyPackageDefinition("a"),
	})

	if err != nil {
		t.Error("Unexpected error")
	}

	if len(pi.packageByName) != len(pi.orderedPackages) {
		t.Error("Index size mismatch")
	}

	packageOrder := packageNames(pi.orderedPackages)
	if packageOrder != "[a b c d]" {
		t.Error("Invalid package order: " + packageOrder)
	}

	checkIndirectDependencies := func(pkgName, expectedDeps string) {
		deps := packageNames(pi.packageByName[pkgName].allReq)
		if deps != expectedDeps {
			t.Error("Indirect dependencies for " + pkgName +
				" do not match: expected=" + expectedDeps +
				"; actual=" + deps)
		}
	}

	checkIndirectDependencies("a", "[]")
	checkIndirectDependencies("b", "[a]")
	checkIndirectDependencies("c", "[a]")
	checkIndirectDependencies("d", "[a b c]")
}
