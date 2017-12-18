// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import (
	"strings"
	"testing"
)

func dummyPackageDefinition(pkgName string) *packageDefinition {
	var pd packageDefinition
	pd.packageName = pkgName
	return &pd
}

func TestDuplicateDefinition(t *testing.T) {
	pkgList := packageDefinitionList{
		dummyPackageDefinition("base"),
		dummyPackageDefinition("client"),
		dummyPackageDefinition("base"),
	}

	pi, err := buildPackageIndex(pkgList)

	if pi != nil || err == nil || !strings.Contains(err.Error(), "dup") {
		t.Error("Package duplicate was not detected")
	}
}
