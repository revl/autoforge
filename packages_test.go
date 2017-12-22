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

func testCircularDependency(t *testing.T,
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
	testCircularDependency(t, packageDefinitionList{
		dummyPackageDefinition("a", "b"),
		dummyPackageDefinition("b", "c"),
		dummyPackageDefinition("c", "a"),
	}, "a -> b -> c -> a")

	testCircularDependency(t, packageDefinitionList{
		dummyPackageDefinition("a", "b"),
		dummyPackageDefinition("b", "c"),
		dummyPackageDefinition("c", "b", "d"),
		dummyPackageDefinition("d"),
	}, "b -> c -> b")

	testCircularDependency(t, packageDefinitionList{
		dummyPackageDefinition("a", "b", "a"),
		dummyPackageDefinition("b"),
	}, "a -> a")
}
