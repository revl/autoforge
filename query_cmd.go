// Copyright (C) 2017 Damon Revoe. All rights reserved.
// Use of this source code is governed by the MIT
// license, which can be found in the LICENSE file.

package main

func queryPackages(workspacedir, pkgpath string) error {
	packageIndex, err := buildPackageIndex(pkgpath)

	if err != nil {
		panic(err)
	}

	packageIndex.printListOfPackages()

	return nil
}
