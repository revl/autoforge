// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

import "fmt"

func queryPackages(workspacedir, pkgpath string) {
	fmt.Println("List of packages:")

	pd := loadPackageDefinition("examples/packages/greeting/greeting.yaml")

	fmt.Println(pd.Name)
}
